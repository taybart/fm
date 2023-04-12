use crossterm::event::{self, Event, KeyEvent};
use std::io;

use crate::log;

use super::render::{setup_tui, teardown_tui};
use super::state::{Action, Mode, State};
use super::tree::Tree;
use super::{command::Command, render::render};

enum InputResult {
    OK,
    Edit,
    Shell,
    Exit,
}

pub struct FM {
    pub tree: Tree,
    pub state: State,
    pub cmd: Command,
}

impl FM {
    pub fn new() -> Result<FM, String> {
        Ok(FM {
            tree: Tree::new()?,
            state: State::default(),
            cmd: Command::default(),
        })
    }

    pub fn run(&mut self) -> io::Result<()> {
        // setup terminal
        let mut terminal = setup_tui()?;

        loop {
            terminal.draw(|f| render(self, f).unwrap_or(()))?;

            if let Event::Key(key) = event::read()? {
                match self.handle_input(key) {
                    InputResult::OK => {}
                    InputResult::Edit => {
                        teardown_tui(&mut terminal)?;
                        if let Err(e) = self.cmd.edit(&mut self.tree, &mut self.state) {
                            log::error(e);
                        }
                        terminal = setup_tui()?;
                        self.state.reset_query();
                    }
                    InputResult::Shell => {
                        teardown_tui(&mut terminal)?;
                        if let Err(e) = self.cmd.shell() {
                            log::error(e);
                        }
                        terminal = setup_tui()?;
                    }
                    InputResult::Exit => {
                        teardown_tui(&mut terminal)?;
                        return Ok(());
                    }
                }
            }
        }
    }

    fn handle_input(&mut self, key: KeyEvent) -> InputResult {
        let show_hidden = self.state.show_hidden;

        match self.state.handle_input(key, &mut self.cmd).action {
            Action::Parent => self.tree.cd_parent(&self.state),
            Action::Selected => {
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
            Action::Up => self.tree.cwd().up(show_hidden),
            Action::Down => self.tree.cwd().down(show_hidden),
            Action::PgUp => self.tree.cwd().pg_up(show_hidden, 10),
            Action::PgDown => self.tree.cwd().pg_down(show_hidden, 10),
            Action::Edit => return InputResult::Edit,
            Action::Shell => return InputResult::Shell,
            Action::Execute => self.cmd.execute(&mut self.tree, &mut self.state),
            Action::SelectFile => {
                self.tree
                    .cwd()
                    .toggle_select_current_file(show_hidden, &self.state.query_string);
                self.tree.cwd().down(show_hidden)
            }
            Action::ResetSelection => self.tree.cwd().state.select(Some(0)),
            Action::Nop => {
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
}
