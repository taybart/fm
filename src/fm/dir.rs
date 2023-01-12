use crate::finder::match_and_score;
use crate::log;

use super::file::File;

use std::fs::{canonicalize, read_dir, read_link};
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
    pub last_selection: usize,
    pub visible_files: Vec<File>,
    pub path: PathBuf,
}

impl Dir {
    pub fn new(dir_name: PathBuf) -> Result<Dir, String> {
        let mut files = Vec::new();
        match read_dir(dir_name.clone()) {
            Ok(dir) => {
                for entry in dir {
                    match File::new(entry.unwrap()) {
                        Ok(file) => {
                            files.push(file);
                        }
                        Err(e) => log::error(e),
                    }
                }

                files.sort_by(|a, b| a.name.cmp(&b.name));

                let path = canonicalize(dir_name).expect("could not canonicalize directory");

                let mut state = ListState::default();
                state.select(Some(0));
                Ok(Dir {
                    state,
                    visible_files: files.clone(),
                    last_selection: 0,
                    files,
                    path,
                })
            }
            Err(e) => Err(format!("could not read dir {dir_name:?}: {e}")),
        }
    }

    pub fn down(&mut self, show_hidden: bool) {
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

    pub fn up(&mut self, show_hidden: bool) {
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

    pub fn pg_down(&mut self, show_hidden: bool, amount: usize) {
        let visible = self.get_visible(show_hidden);
        let i = match self.state.selected() {
            Some(i) => {
                self.last_selection = i;
                let update = i + amount;
                if update < visible.len() - 1 {
                    update
                } else {
                    visible.len() - 1
                }
            }
            None => 0,
        };
        self.state.select(Some(i));
    }

    pub fn pg_up(&mut self, _show_hidden: bool, amount: usize) {
        let i = match self.state.selected() {
            Some(i) => {
                self.last_selection = i;
                let update = i as i32 - amount as i32;
                if update >= 0 {
                    i - amount
                } else {
                    0
                }
            }
            None => 0,
        };
        self.state.select(Some(i));
    }

    pub fn ensure_selection(&mut self, show_hidden: bool, query: &str) {
        let visible = self.get_visible_with_query(show_hidden, query);

        if visible.is_empty() {
            return self.state.select(Some(0));
        }
        let i = match self.state.selected() {
            Some(i) => {
                // self.last_selection = i;
                if i >= visible.len() {
                    visible.len() - 1
                } else {
                    i
                }
            }
            None => 0,
        };
        self.state.select(Some(i));
    }

    pub fn _mark_selected(&self, _name: String, _show_hidden: bool) -> usize {
        0
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
        self.files
            .clone()
            .into_iter()
            .filter(|x| if !show_hidden { !x.is_hidden } else { true })
            .filter(|x| match_and_score(query, &x.name).is_some())
            .collect::<Vec<File>>()
    }
    fn get_visible(&self, show_hidden: bool) -> Vec<File> {
        self.files
            .clone()
            .into_iter()
            .filter(|x| if !show_hidden { !x.is_hidden } else { true })
            .collect::<Vec<File>>()
    }

    pub fn get_selected_file(&self, show_hidden: bool, query: &str) -> Option<File> {
        let visible = self.get_visible_with_query(show_hidden, query);
        match self.state.selected() {
            Some(i) => visible.get(i).cloned(),
            None => None,
        }
    }

    pub fn toggle_select_current_file(&mut self, show_hidden: bool, query: &str) {
        let mut visible: Vec<&mut File> = self
            .files
            .iter_mut()
            .filter(|x| if !show_hidden { !x.is_hidden } else { true })
            .filter(|x| match_and_score(query, &x.name).is_some())
            .collect::<Vec<&mut File>>();

        if let Some(i) = self.state.selected() {
            let file = visible.get_mut(i).unwrap();
            file.toggle_selected();
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
                if file.selected {
                    style = style.add_modifier(Modifier::BOLD);
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
                let mut name = f.name.to_owned();
                let mut style = Style::default();
                if f.is_dir {
                    style = style.fg(Color::LightBlue);
                }
                if f.is_symlink {
                    if let Ok(path) = read_link(&name) {
                        name += format!("~>{}", path.to_string_lossy()).as_str();
                    } else {
                        name += "~>?";
                    }
                    style = style.fg(Color::Red);
                }
                ListItem::new(name).style(style)
            })
            .collect();
        let list = List::new(files)
            .highlight_style(Style::default().add_modifier(Modifier::BOLD | Modifier::UNDERLINED));

        if stateful {
            f.render_stateful_widget(list, rect, &mut self.state);
        } else {
            f.render_widget(list.style(Style::default()), rect);
        }
    }
}
