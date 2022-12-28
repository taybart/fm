use crate::finder::match_and_score;

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

#[derive(Clone)]
pub struct Dir {
    pub state: ListState,
    pub files: Vec<File>,
    pub visible_files: Vec<File>,
    // pub show_hidden: bool,
    pub path: PathBuf,
}

impl Dir {
    pub fn new(dir_name: PathBuf) -> Result<Dir, String> {
        let mut files = Vec::new();
        match read_dir(dir_name.clone()) {
            Ok(dir) => {
                // should sort dir > files
                for entry in dir {
                    let file = File::new(entry.unwrap());
                    files.push(file);
                }

                files.sort_by(|a, b| a.name.cmp(&b.name));

                let path = canonicalize(dir_name).expect("could not canonicalize directory");

                let mut state = ListState::default();
                state.select(Some(0));
                Ok(Dir {
                    state,
                    visible_files: files.clone(),
                    files,
                    path,
                })
            }
            Err(e) => Err(format!("could not read dir {dir_name:?}: {e}")),
        }
    }

    pub fn next(&mut self, show_hidden: bool) {
        let visible = self.get_visible(show_hidden);
        let i = match self.state.selected() {
            Some(i) => {
                if i >= visible.len() - 1 {
                    0
                } else {
                    i + 1
                }
            }
            None => 0,
        };
        self.state.select(Some(i));
    }

    pub fn previous(&mut self, show_hidden: bool) {
        let visible = self.get_visible(show_hidden);
        let i = match self.state.selected() {
            Some(i) => {
                if i == 0 {
                    visible.len() - 1
                } else {
                    i - 1
                }
            }
            None => 0,
        };
        self.state.select(Some(i));
    }

    pub fn ensure_selection(&mut self, show_hidden: bool, query: &str) {
        let visible = self.get_visible_with_query(show_hidden, query);
        // crate::log::write(format!("{visible:?}"));
        let i = match self.state.selected() {
            Some(i) => {
                if i >= visible.len() {
                    visible.len() - 1
                } else {
                    i
                }
            }
            None => 0,
        };
        // crate::log::write(format!("idx {i}"));
        self.state.select(Some(i));
    }

    pub fn index_by_name(&self, name: String, show_hidden: bool) -> usize {
        let visible = self.get_visible(show_hidden);
        for (i, f) in visible.iter().enumerate() {
            if name == f.name {
                return i;
            }
        }
        0
    }

    fn get_visible_with_query(&self, show_hidden: bool, query: &str) -> Vec<File> {
        let files = self.files.clone();

        let hidden = files
            .into_iter()
            .filter(|x| if !show_hidden { !x.is_hidden } else { true })
            .collect::<Vec<File>>();

        let mut visible = vec![];
        for file in hidden.into_iter() {
            if let Some(_matches) = match_and_score(query, &file.name) {
                visible.push(file);
            }
        }
        visible
    }
    fn get_visible(&self, show_hidden: bool) -> Vec<File> {
        let visible = self.files.clone();

        visible
            .into_iter()
            .filter(|x| if !show_hidden { !x.is_hidden } else { true })
            .collect::<Vec<File>>()
    }

    pub fn get_selected_file(&self, show_hidden: bool, query: &str) -> Option<File> {
        let visible = self.get_visible_with_query(show_hidden, query);
        match self.state.selected() {
            Some(i) => visible.into_iter().nth(i),
            None => None,
        }
    }

    pub fn render_with_query<B: Backend>(
        &mut self,
        f: &mut Frame<B>,
        query: &str,
        rect: Rect,
        show_hidden: bool,
    ) {
        let mut files: Vec<(f64, ListItem)> = Vec::new();

        let visible = self.get_visible_with_query(show_hidden, query);
        for file in visible.iter() {
            if let Some((score, span)) = file.display_with_query(query) {
                let mut style = Style::default();
                if file.is_dir {
                    style = style.fg(Color::LightBlue);
                }
                files.push((score, ListItem::new(span).style(style)));
            }
        }

        // sort by scores
        files.sort_by(|a, b| b.0.partial_cmp(&a.0).unwrap());

        let list = List::new(
            files
                .iter()
                .map(|l| l.1.to_owned())
                .collect::<Vec<ListItem>>(),
        )
        .highlight_style(Style::default().add_modifier(Modifier::BOLD | Modifier::UNDERLINED));
        // .highlight_symbol(">");

        f.render_stateful_widget(list, rect, &mut self.state);
    }

    pub fn render<B: Backend>(
        &mut self,
        f: &mut Frame<B>,
        stateful: bool,
        rect: Rect,
        show_hidden: bool,
    ) {
        let files: Vec<ListItem> = self
            .files
            .iter()
            .filter(|f| !f.is_hidden || show_hidden)
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
            f.render_widget(list.style(Style::default()), rect);
        }
    }
}
