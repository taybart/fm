use super::dir::Dir;
use super::state::{Command, Mode, State};
use std::{
    collections::HashMap,
    env, fs, io,
    path::{Path, PathBuf},
    process::Command as cmd,
};

use tui::{
    backend::{Backend, CrosstermBackend},
    layout::{Constraint, Direction, Layout, Rect},
    style::{Modifier, Style},
    text::Text,
    widgets::Paragraph,
    Frame, Terminal,
};

use crossterm::{
    event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyEvent},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};

enum InputResult {
    OK,
    Edit,
    Exit,
}

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

    pub fn edit(&mut self) -> Result<(), String> {
        let editor = std::env::var("EDITOR").map_err(|e| format!("could not get editor {e}"))?;

        let query = self.state.query_string.clone();
        let show_hidden = self.state.show_hidden;

        let file = self.cwd().get_selected_file(show_hidden, &query).unwrap();
        if !file.is_dir {
            let mut child = cmd::new(editor)
                .arg(file.name)
                .spawn()
                .map_err(|e| format!("failed to start editor {e}"))?;

            child.wait().map_err(|e| format!("child failed {e}"))?;
        }
        Ok(())
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
        let query = self.state.query_string.clone();
        let show_hidden = self.state.show_hidden;
        if let Some(selected) = self.cwd().get_selected_file(show_hidden, &query) {
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
        let query = self.state.query_string.clone();
        /*
         * Left column
         */
        self.parent().render(f, false, chunks[0], show_hidden);

        /*
         * middle column
         */
        match self.state.mode {
            Mode::Search => {
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

        if let Some(selected) = self.cwd().get_selected_file(show_hidden, &query) {
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

    pub fn run(&mut self) -> io::Result<()> {
        // setup terminal
        enable_raw_mode()?;
        let mut stdout = io::stdout();
        execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
        let backend = CrosstermBackend::new(stdout);
        let mut terminal = Terminal::new(backend)?;

        loop {
            terminal.draw(|f| self.render(f).unwrap_or(()))?;

            if let Event::Key(key) = event::read()? {
                match self.handle_input(key) {
                    InputResult::OK => {}
                    InputResult::Edit => {
                        // restore terminal
                        disable_raw_mode()?;
                        execute!(
                            terminal.backend_mut(),
                            LeaveAlternateScreen,
                            DisableMouseCapture
                        )?;
                        terminal.show_cursor()?;
                        if let Err(e) = self.edit() {
                            crate::log::error(e);
                        }
                        // setup terminal
                        enable_raw_mode()?;
                        let mut stdout = io::stdout();
                        execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
                        let backend = CrosstermBackend::new(stdout);
                        terminal = Terminal::new(backend)?;
                    }
                    InputResult::Exit => {
                        // restore terminal
                        disable_raw_mode()?;
                        execute!(
                            terminal.backend_mut(),
                            LeaveAlternateScreen,
                            DisableMouseCapture
                        )?;
                        return Ok(());
                    }
                }
            }
        }
    }

    fn handle_input(&mut self, key: KeyEvent) -> InputResult {
        let show_hidden = self.state.show_hidden;
        let query = self.state.query_string.clone();

        match self.state.handle_input(key).command {
            Command::Parent => self.cd_parent(),
            Command::Selected => {
                if let Some(selected) = self.cwd().get_selected_file(show_hidden, &query) {
                    if selected.is_dir {
                        self.cd_selected();
                    } else {
                        return InputResult::Edit;
                    }
                    self.state.reset_command();
                }
            }
            Command::Up => self.cwd().next(show_hidden),
            Command::Down => self.cwd().previous(show_hidden),
            Command::Edit => return InputResult::Edit,
            Command::Nop => {
                if self.state.mode == Mode::Search {
                    let query = self.state.query_string.clone();
                    self.cwd().ensure_selection(show_hidden, &query);
                }
            }
        };
        if self.state.exit {
            InputResult::Exit
        } else {
            InputResult::OK
        }
    }
}
