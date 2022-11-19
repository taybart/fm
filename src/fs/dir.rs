use crate::fs::file::File;

use std::fs::read_dir;

use tui::{
    style::{Color, Modifier, Style},
    widgets::{List, ListItem, ListState},
};

pub struct Dir {
    pub state: ListState,
    pub files: Vec<File>,
    pub show_hidden: bool,
}

impl Dir {
    pub fn new(dir_name: &str) -> Dir {
        let mut files = Vec::new();
        // should sort dir > files
        for entry in read_dir(dir_name).unwrap() {
            files.push(File::new(entry.unwrap()));
            // do sort?
        }

        let mut state = ListState::default();
        state.select(Some(0));
        Dir {
            state,
            files,
            show_hidden: false,
        }
    }

    // fn sort(&mut self) {
    // }

    pub fn next(&mut self) {
        let i = match self.state.selected() {
            Some(i) => {
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

    pub fn get_selected_contents(&self) -> String {
        match self.state.selected() {
            Some(i) => self.files.get(i).unwrap().get_contents(),
            None => "".to_string(),
        }
    }

    pub fn to_tui_list(&self) -> List<'static> {
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
        List::new(files).highlight_style(
            Style::default()
                .bg(Color::LightBlue)
                .add_modifier(Modifier::BOLD),
        )
    }
}
