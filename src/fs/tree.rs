use super::dir::Dir;
use std::{
    collections::HashMap,
    env, fs,
    path::{Path, PathBuf},
};

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
    fs_tree: HashMap<PathBuf, Dir>,
    pub show_hidden: bool,
    pub mode: Mode,
    pub query: String,
}

impl Tree {
    pub fn new() -> Tree {
        let show_hidden = true;
        let parent = Dir::new(Tree::parent_path().expect("parent path idk")).unwrap();
        let cwd = Dir::new(Tree::cwd_path().expect("cwd path idk")).unwrap();

        let mut fs_tree: HashMap<PathBuf, Dir> = HashMap::new();
        fs_tree.insert(parent.path.clone(), parent.clone());
        fs_tree.insert(cwd.path.clone(), cwd.clone());

        Tree {
            fs_tree,
            mode: Mode::NORMAL,
            query: String::new(),
            show_hidden,
        }
    }

    pub fn parent_path() -> Result<PathBuf, std::io::Error> {
        fs::canonicalize(Path::new(".."))
    }
    pub fn cwd_path() -> Result<PathBuf, std::io::Error> {
        fs::canonicalize(Path::new("."))
    }

    pub fn parent(&mut self) -> &mut Dir {
        let parent_path = Tree::parent_path().expect("could not canonicalize parent path");
        self.fs_tree.get_mut(&parent_path).unwrap()
    }
    pub fn cwd(&mut self) -> &mut Dir {
        let cwd_path = fs::canonicalize(Path::new(".")).expect("could not canonicalize cwd path");
        self.fs_tree.get_mut(&cwd_path).unwrap()
    }

    pub fn cd_parent(&mut self) {
        let init_parent_path = Tree::cwd_path().expect("could not canonicalize parent path");

        env::set_current_dir(Path::new("..")).unwrap();
        let parent_path = Tree::parent_path().expect("could not canonicalize parent path");

        self.fs_tree
            .entry(parent_path.clone())
            .or_insert(Dir::new(parent_path).unwrap());
        let cwd_path = Tree::cwd_path().expect("could not canonicalize cwd path");
        self.fs_tree
            .entry(cwd_path.clone())
            .or_insert(Dir::new(cwd_path.clone()).unwrap());
        let cwd = self.fs_tree.get_mut(&cwd_path).unwrap();
        let idx = cwd.index_by_name(
            init_parent_path
                .file_name()
                .unwrap()
                .to_os_string()
                .into_string()
                .unwrap(),
            self.show_hidden,
        );
        cwd.state.select(Some(idx));
    }
    pub fn cd_selected(&mut self) {
        let show_hidden = self.show_hidden;
        match self.cwd().get_selected_file(show_hidden) {
            Some(selected) => {
                if selected.is_dir {
                    env::set_current_dir(Path::new(&selected.name)).unwrap();
                    let cwd_path =
                        fs::canonicalize(Path::new(".")).expect("could not canonicalize cwd path");

                    self.fs_tree
                        .entry(cwd_path.clone())
                        .or_insert(Dir::new(cwd_path).unwrap());
                }
            }
            None => {}
        }
    }

    pub fn toggle_show_hidden(&mut self) {
        self.show_hidden = !self.show_hidden;
    }

    pub fn render<B: Backend>(&mut self, f: &mut Frame<B>) -> Result<(), String> {
        let chunks = Layout::default()
            .direction(Direction::Horizontal)
            .vertical_margin(1)
            .horizontal_margin(3)
            .constraints(
                [
                    Constraint::Percentage(20),
                    Constraint::Percentage(30),
                    Constraint::Percentage(50),
                ]
                .as_ref(),
            )
            .split(f.size());

        let show_hidden = self.show_hidden;
        self.parent().render(f, false, chunks[0], show_hidden);
        match self.mode {
            Mode::SEARCH => {
                let query = self.query.clone();
                self.cwd()
                    .render_with_query(f, &query, chunks[1], show_hidden);
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
            }
            Mode::NORMAL => {
                self.cwd().render(f, true, chunks[1], show_hidden);
            }
        };

        match self.cwd().get_selected_file(show_hidden) {
            Some(selected) => {
                if selected.is_dir {
                    // TODO: render messsage about perissions if that fails
                    Dir::new(selected.path.clone())?.render(f, false, chunks[2], show_hidden);
                } else {
                    f.render_widget(
                        Paragraph::new(Text::raw(selected.get_contents())),
                        chunks[2],
                    );
                }
            }
            None => {}
        };
        Ok(())
    }
}
