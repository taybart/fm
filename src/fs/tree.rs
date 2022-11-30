use super::dir::Dir;
use std::{collections::HashMap, env, path::Path};

use tui::{
    backend::Backend,
    layout::{Constraint, Direction, Layout, Rect},
    style::{Modifier, Style},
    text::Text,
    widgets::Paragraph,
    Frame,
};

#[derive(PartialEq)]
pub enum Mode {
    NORMAL,
    SEARCH,
}

pub struct Tree {
    fs_tree: HashMap<String, Dir>,
    pub show_hidden: bool,
    pub mode: Mode,
    pub parent: Dir,
    pub cwd: Dir,
    pub query: String,
}

impl Tree {
    pub fn new() -> Tree {
        let show_hidden = false;
        let parent = Dir::new("..", show_hidden);
        let cwd = Dir::new(".", show_hidden);
        Tree {
            fs_tree: HashMap::new(),
            mode: Mode::NORMAL,
            query: String::new(),
            show_hidden,
            parent,
            cwd,
        }
    }
    pub fn cd_up(&mut self) {
        // TODO: check if dir exists in self.fs_tree
        env::set_current_dir(Path::new("..")).unwrap();
        self.parent = Dir::new("..", self.show_hidden);
        self.cwd = Dir::new(".", self.show_hidden);
    }
    pub fn cd_down(&mut self) {
        // TODO: check if dir exists in self.fs_tree
        match self.cwd.get_selected_file() {
            Some(selected) => {
                if selected.is_dir {
                    env::set_current_dir(Path::new(&selected.name)).unwrap();
                    self.parent = Dir::new("..", self.show_hidden);
                    self.cwd = Dir::new(".", self.show_hidden);
                }
            }
            None => {}
        }
    }

    pub fn toggle_show_hidden(&mut self) {
        self.show_hidden = !self.show_hidden;

        self.cwd = Dir::new(".", self.show_hidden);
        self.parent = Dir::new("..", self.show_hidden);
    }

    pub fn render<B: Backend>(&mut self, f: &mut Frame<B>) {
        let chunks = Layout::default()
            .direction(Direction::Horizontal)
            .margin(1)
            .constraints(
                [
                    Constraint::Percentage(20),
                    Constraint::Percentage(30),
                    Constraint::Percentage(50),
                ]
                .as_ref(),
            )
            .split(f.size());

        self.parent.render(f, false, chunks[0]);
        if self.mode == Mode::SEARCH {
            // f.render_widget(self.cwd.display_with_query(&self.query.clone()), chunks[1]);
            self.cwd
                .render_with_query(f, &self.query.clone(), chunks[1]);
            f.render_widget(
                Paragraph::new(Text::raw(format!("> {}", &self.query.clone())))
                    .style(Style::default().add_modifier(Modifier::UNDERLINED)),
                Rect::new(chunks[1].x + 1, f.size().height - 1, chunks[1].width, 1),
            );
            f.set_cursor(
                // TODO: add constants for offsets?
                // Put cursor past the end of the input text
                chunks[1].x + self.query.len() as u16 + 3,
                f.size().height,
            )
        } else {
            self.cwd.render(f, true, chunks[1]);
        }

        match self.cwd.get_selected_file() {
            Some(selected) => {
                if selected.is_dir {
                    Dir::new(selected.name.as_str(), self.show_hidden).render(f, false, chunks[2]);
                } else {
                    f.render_widget(
                        Paragraph::new(Text::raw(selected.get_contents())),
                        chunks[2],
                    );
                }
            }
            None => {}
        };
    }
}
