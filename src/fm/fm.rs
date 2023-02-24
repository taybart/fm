use super::dir::Dir;
use super::state::{Command, Mode, State};
use super::tree::Tree;
use crate::log;
use std::{io, process::Command as cmd};

use tui::{
    backend::{Backend, CrosstermBackend},
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
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
    Shell,
    Exit,
}

pub struct FM {
    tree: Tree,
    pub state: State,
}

impl FM {
    pub fn new() -> Result<FM, String> {
        Ok(FM {
            tree: Tree::new()?,
            state: State::default(),
        })
    }

    pub fn shell(&mut self) -> Result<(), String> {
        let shell = std::env::var("SHELL").map_err(|e| format!("could not get editor {e}"))?;

        let mut child = cmd::new(shell)
            .spawn()
            .map_err(|e| format!("failed to start editor {e}"))?;

        child.wait().map_err(|e| format!("child failed {e}"))?;
        Ok(())
    }

    pub fn edit(&mut self) -> Result<(), String> {
        // TODO: better error handling
        let editor = std::env::var("EDITOR").map_err(|e| format!("could not get editor {e}"))?;

        let file = self
            .tree
            .cwd()
            .get_selected_file(self.state.show_hidden, &self.state.query_string)
            .unwrap();
        if !file.is_dir {
            let mut child = cmd::new(editor)
                .arg(file.name)
                .spawn()
                .map_err(|e| format!("failed to start editor {e}"))?;

            child.wait().map_err(|e| format!("child failed {e}"))?;
        }
        Ok(())
    }

    pub fn render<B: Backend>(&mut self, f: &mut Frame<B>) -> Result<(), String> {
        let hide_parent = self.state.hide_parent;
        // TODO: add to config
        if f.size().width < 80 {
            self.state.hide_parent = true;
        }

        let constraints = if self.state.hide_parent {
            vec![Constraint::Percentage(30), Constraint::Percentage(70)]
        } else {
            vec![
                Constraint::Percentage(20),
                Constraint::Percentage(30),
                Constraint::Percentage(50),
            ]
        };

        let chunks = Layout::default()
            .direction(Direction::Horizontal)
            .vertical_margin(1)
            .horizontal_margin(3)
            .constraints(constraints)
            .split(f.size());

        f.render_widget(
            Paragraph::new(Text::raw(Tree::cwd_path()?.to_str().unwrap_or(""))).style(
                Style::default()
                    .add_modifier(Modifier::BOLD)
                    .fg(Color::LightGreen),
            ),
            Rect::new(1, 0, f.size().width - 1, 1),
        );

        let show_hidden = self.state.show_hidden;
        /*
         * Left column
         */
        if !self.state.hide_parent {
            self.tree.parent().render(f, false, chunks[0], show_hidden);
        }

        /*
         * middle column
         */
        let idx = if self.state.hide_parent { 0 } else { 1 };
        match self.state.mode {
            Mode::Search => {
                self.tree.cwd().render_with_query(
                    f,
                    &self.state.query_string,
                    chunks[idx],
                    show_hidden,
                );
                f.render_widget(
                    Paragraph::new(Text::raw(format!(">{}", &self.state.query_string)))
                        .style(Style::default().add_modifier(Modifier::UNDERLINED)),
                    Rect::new(chunks[idx].x, f.size().height - 1, chunks[idx].width, 1),
                );
                f.set_cursor(
                    chunks[idx].x + self.state.query_string.len() as u16 + 1,
                    f.size().height,
                )
            }
            Mode::Normal => {
                self.tree.cwd().render(f, true, chunks[idx], show_hidden);
            }
            Mode::Command => {
                self.tree.cwd().render(f, true, chunks[idx], show_hidden);
                f.render_widget(
                    Paragraph::new(Text::raw(format!(":{}", &self.state.command_string)))
                        .style(Style::default().add_modifier(Modifier::UNDERLINED)),
                    Rect::new(chunks[0].x, f.size().height - 1, chunks[idx].width, 1),
                );
                f.set_cursor(
                    chunks[0].x + self.state.command_string.len() as u16 + 1,
                    f.size().height,
                )
            }
        };

        /*
         * right column
         */

        let idx = if self.state.hide_parent { 1 } else { 2 };
        if let Some(selected) = self
            .tree
            .cwd()
            .get_selected_file(show_hidden, &self.state.query_string)
        {
            if selected.is_dir {
                Dir::new(selected.path)?.render(f, false, chunks[idx], show_hidden);
            } else {
                f.render_widget(Paragraph::new(Text::raw(selected.contents())), chunks[idx]);
            }
        };

        self.state.hide_parent = hide_parent;
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
                            log::error(e);
                        }
                        // setup terminal
                        enable_raw_mode()?;
                        let mut stdout = io::stdout();
                        execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
                        let backend = CrosstermBackend::new(stdout);
                        terminal = Terminal::new(backend)?;
                        self.state.reset_query();
                    }
                    InputResult::Shell => {
                        // restore terminal
                        disable_raw_mode()?;
                        execute!(
                            terminal.backend_mut(),
                            LeaveAlternateScreen,
                            DisableMouseCapture
                        )?;
                        terminal.show_cursor()?;
                        if let Err(e) = self.shell() {
                            log::error(e);
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

        match self.state.handle_input(key).command {
            Command::Parent => self.tree.cd_parent(&self.state),
            Command::Selected => {
                if let Some(selected) = self
                    .tree
                    .cwd()
                    .get_selected_file(show_hidden, &self.state.query_string)
                {
                    if selected.is_dir {
                        self.tree.cd_selected(&self.state);
                        self.state.reset_query();
                    } else if !self.state.query_string.is_empty() {
                        log::write(format!("edit {}", selected.name));
                        return InputResult::Edit;
                    }
                }
            }
            Command::Up => self.tree.cwd().up(show_hidden),
            Command::Down => self.tree.cwd().down(show_hidden),
            Command::PgUp => self.tree.cwd().pg_up(show_hidden, 10),
            Command::PgDown => self.tree.cwd().pg_down(show_hidden, 10),
            Command::Edit => return InputResult::Edit,
            Command::Shell => return InputResult::Shell,
            Command::Execute => self.execute_command(),
            Command::SelectFile => {
                self.tree
                    .cwd()
                    .toggle_select_current_file(show_hidden, &self.state.query_string);
                self.tree.cwd().down(show_hidden)
            }
            Command::ResetSelection => self.tree.cwd().state.select(Some(0)),
            Command::Nop => {
                if self.state.mode == Mode::Search {
                    self.tree
                        .cwd()
                        .ensure_selection(show_hidden, &self.state.query_string);
                }
            }
        };
        if self.state.exit {
            InputResult::Exit
        } else {
            InputResult::OK
        }
    }

    fn execute_command(&mut self) {
        // | :delete       | ed      | Moves file to temporary location. After fm is closed, the files will be deleted permanently            |
        // | :undo         | eu      | Put files back where they were and don't delete them at the end of the fm session.                     |
        // | :yank         | yy      | Copy file under cursor                                                                                 |
        // | :cut          | dd      | Cut file under cursor                                                                                  |
        // | :paste        | pp      | Paste file to current directory                                                                        |

        let cmds = self.state.command_string.split(' ').collect::<Vec<&str>>();
        match cmds[0] {
            "rename" | "rn" => {
                // TODO: if no cmd[1] ask for name
                if let Some(new_name) = cmds.get(1) {
                    self.tree
                        .cwd()
                        .rename_selected(new_name, self.state.show_hidden);
                }
            }
            "cd" => {
                log::write(format!("cd {}", cmds[1]));
                let dir = shellexpand::tilde(cmds[1]);
                let dir = std::fs::canonicalize(dir.into_owned()).expect("idk");
                if let Err(e) = std::env::set_current_dir(&dir) {
                    log::error(format!("unknown command {e}"));
                }
                self.tree.cd(dir);
            }
            "exit" | "q" => {
                log::write("quitting".to_string());
                self.state.exit();
            }
            "edit" | "e" => {
                log::write("editing".to_string());
                self.edit().expect("could not edit");
            }
            "hidden" | "h" => {
                log::write("toggle hidden".to_string());
                self.state.toggle_hidden();
            }
            _ => match cmds[0].chars().nth(0).unwrap() {
                '!' => {
                    let mut exec = cmds.get(0).unwrap().to_string();
                    exec.remove(0);
                    let args = cmds.get(1..).unwrap();
                    log::write(format!("execute {} {:?}", exec, args));

                    match cmd::new(exec).args(args).spawn() {
                        Ok(mut child) => {
                            if let Err(e) = child.wait() {
                                log::error(e.to_string())
                            }
                        }
                        Err(e) => log::error(e.to_string()),
                    }

                    if let Err(e) = self.tree.cwd().refresh() {
                        log::error(e);
                    }
                }
                _ => {
                    log::error(format!("unknown command {:?}", cmds));
                }
            },
        }
        self.state.reset_command();
    }
}
