use super::dir::Dir;
use super::state::{Command, Mode, State};
use std::{
    collections::HashMap,
    env, fs,
    path::{Path, PathBuf},
};

use crossterm::event::KeyEvent;

use tui::{
    backend::Backend,
    layout::{Constraint, Direction, Layout, Rect},
    style::{Modifier, Style},
    text::Text,
    widgets::Paragraph,
    Frame,
};

pub struct Tree {
    fs_tree: HashMap<PathBuf, Dir>,
    pub state: State,
}

impl Tree {
    pub fn new() -> Tree {
        let parent = Dir::new(Tree::parent_path().expect("parent path idk")).unwrap();
        let cwd = Dir::new(Tree::cwd_path().expect("cwd path idk")).unwrap();

        let mut fs_tree: HashMap<PathBuf, Dir> = HashMap::new();
        fs_tree.insert(parent.path.clone(), parent);
        fs_tree.insert(cwd.path.clone(), cwd);

        let state = State::default();

        Tree { fs_tree, state }
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
            .or_insert_with(|| Dir::new(parent_path).unwrap());
        let cwd_path = Tree::cwd_path().expect("could not canonicalize cwd path");
        self.fs_tree
            .entry(cwd_path.clone())
            .or_insert_with(|| Dir::new(cwd_path.clone()).unwrap());
        let cwd = self.fs_tree.get_mut(&cwd_path).unwrap();
        let idx = cwd.index_by_name(
            init_parent_path
                .file_name()
                .unwrap()
                .to_os_string()
                .into_string()
                .unwrap(),
            self.state.show_hidden,
        );
        cwd.state.select(Some(idx));
    }
    pub fn cd_selected(&mut self) {
        let show_hidden = self.state.show_hidden;
        if let Some(selected) = self.cwd().get_selected_file(show_hidden) {
            if selected.is_dir {
                env::set_current_dir(Path::new(&selected.name)).unwrap();
                let cwd_path =
                    fs::canonicalize(Path::new(".")).expect("could not canonicalize cwd path");

                self.fs_tree
                    .entry(cwd_path.clone())
                    .or_insert_with(|| Dir::new(cwd_path).unwrap());
            }
        }
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

        let show_hidden = self.state.show_hidden;
        /*
         * Left column
         */
        self.parent().render(f, false, chunks[0], show_hidden);

        /*
         * middle column
         */
        match self.state.mode {
            Mode::Search => {
                let query = self.state.command_string.clone();
                self.cwd()
                    .render_with_query(f, &query, chunks[1], show_hidden);
                f.render_widget(
                    Paragraph::new(Text::raw(format!("> {}", &query.clone())))
                        .style(Style::default().add_modifier(Modifier::UNDERLINED)),
                    Rect::new(chunks[1].x + 1, f.size().height - 1, chunks[1].width, 1),
                );
                f.set_cursor(
                    // TODO: add constants for offsets?
                    // Put cursor past the end of the input text
                    chunks[1].x + query.len() as u16 + 3,
                    f.size().height,
                )
            }
            Mode::Normal => {
                self.cwd().render(f, true, chunks[1], show_hidden);
            }
            Mode::Command => {
                f.render_widget(
                    Paragraph::new(Text::raw(format!(
                        ":{}",
                        &self.state.command_string.clone()
                    )))
                    .style(Style::default().add_modifier(Modifier::UNDERLINED)),
                    Rect::new(chunks[0].x + 1, f.size().height - 1, chunks[1].width, 1),
                );
                f.set_cursor(
                    // TODO: add constants for offsets?
                    // Put cursor past the end of the input text
                    chunks[0].x + self.state.command_string.len() as u16 + 2,
                    f.size().height,
                )
            }
        };

        /*
         * right column
         */
        if let Some(selected) = self.cwd().get_selected_file(show_hidden) {
            if selected.is_dir {
                // TODO: render messsage about perissions if that fails
                Dir::new(selected.path)?.render(f, false, chunks[2], show_hidden);
            } else {
                f.render_widget(
                    Paragraph::new(Text::raw(selected.get_contents())),
                    chunks[2],
                );
            }
        };
        Ok(())
    }

    pub fn handle_input(&mut self, key: KeyEvent) -> bool {
        let show_hidden = self.state.show_hidden;
        match self.state.handle_input(key).command {
            Command::Parent => self.cd_parent(),
            Command::Selected => self.cd_selected(),
            Command::Up => self.cwd().next(show_hidden),
            Command::Down => self.cwd().previous(show_hidden),
            Command::Nop => {
                return self.state.exit;
            }
        };
        self.state.exit
    }
}
