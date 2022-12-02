use super::file::File;

use std::fs::{canonicalize, read_dir};
use std::path::PathBuf;

use tui::{
    backend::Backend,
    layout::Rect,
    style::{Color, Modifier, Style},
    widgets::{List, ListItem, ListState},
    Frame,
};

// pub struct DirState {
//     pub path: PathBuf,
//     pub selected: usize,
//     pub files: Vec<File>,
// }

#[derive(Clone)]
pub struct Dir {
    pub state: ListState,
    pub files: Vec<File>,
    pub visible_files: Vec<File>,
    pub show_hidden: bool,
    pub path: PathBuf,
}

impl Dir {
    pub fn new(dir_name: &str, show_hidden: bool) -> Dir {
        let mut files = Vec::new();
        let dir = read_dir(dir_name).unwrap();
        // should sort dir > files
        for entry in dir {
            let file = File::new(entry.unwrap());
            if file.is_hidden && show_hidden {
                files.push(file);
            } else {
                files.push(file);
            }
        }

        files.sort_by(|a, b| a.name.cmp(&b.name));

        let path = canonicalize(PathBuf::from(dir_name)).expect("could not canonicalize directory");

        let mut state = ListState::default();
        state.select(Some(0));
        // let mut search_state = ListState::default();
        // search_state.select(Some(0));
        Dir {
            state,
            // search_state,
            visible_files: files.clone(),
            files,
            show_hidden,
            path,
        }
    }

    pub fn next(&mut self) {
        let i = match self.state.selected() {
            Some(i) => {
                // TODO: selection makes no sense
                /* if i < self.files.len() - 1 {
                    let mut inc = i + 1;
                    while !self.show_hidden && self.files.get(inc).unwrap().is_hidden {
                        inc += 1;
                    }
                    inc
                } else {
                    0
                } */
                if i >= self.files.len() - 1 {
                    0
                } else {
                    i + 1
                }
            }
            None => 0,
        };
        self.state.select(Some(i));
    }

    pub fn previous(&mut self) {
        let i = match self.state.selected() {
            Some(i) => {
                if i == 0 {
                    self.files.len() - 1
                } else {
                    i - 1
                }
            }
            None => 0,
        };
        self.state.select(Some(i));
    }

    pub fn get_selected_file(&self) -> Option<&File> {
        match self.state.selected() {
            Some(i) => self.files.get(i),
            None => None,
        }
    }

    pub fn render_with_query<B: Backend>(&mut self, f: &mut Frame<B>, query: &str, rect: Rect) {
        let mut lines: Vec<(f64, ListItem)> = Vec::new();
        for file in &self.files {
            match file.display_with_query(query) {
                Some((score, span)) => {
                    let mut style = Style::default();
                    if file.is_dir {
                        style = style.fg(Color::LightBlue);
                    }
                    lines.push((score, ListItem::new(span).style(style)));
                }
                None => {}
            }
        }

        // sort by scores
        lines.sort_by(|a, b| b.0.partial_cmp(&a.0).unwrap());

        let list = List::new(
            lines
                .iter()
                .map(|l| l.1.to_owned())
                .collect::<Vec<ListItem>>(),
        )
        .highlight_style(Style::default().add_modifier(Modifier::BOLD | Modifier::UNDERLINED));
        // .highlight_symbol(">");

        f.render_stateful_widget(list, rect, &mut self.state);
    }

    pub fn render<B: Backend>(&mut self, f: &mut Frame<B>, stateful: bool, rect: Rect) {
        let files: Vec<ListItem> = self
            .files
            .iter()
            .filter(|f| !f.is_hidden || self.show_hidden)
            .map(|f| {
                let mut style = Style::default();
                if f.is_dir {
                    style = style.fg(Color::LightBlue);
                }
                ListItem::new(f.name.to_owned()).style(style)
            })
            .collect();
        let list = List::new(files)
            .highlight_style(Style::default().add_modifier(Modifier::BOLD | Modifier::UNDERLINED));
        // .highlight_symbol(">");

        if stateful {
            f.render_stateful_widget(list, rect, &mut self.state);
        } else {
            f.render_widget(list, rect);
        }
    }
}
