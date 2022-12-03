use crate::fs::tree::{Mode, Tree};

mod finder;
mod fs;

use crossterm::{
    event::{self, DisableMouseCapture, EnableMouseCapture, Event, KeyCode, KeyModifiers},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use std::{error::Error, io};
use tui::{
    backend::{Backend, CrosstermBackend},
    Terminal,
};

fn main() -> Result<(), Box<dyn Error>> {
    // setup terminal
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    // create app and run it
    let app = Tree::new();
    let res = run_app(&mut terminal, app);

    // restore terminal
    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;

    if let Err(err) = res {
        println!("{:?}", err)
    }

    Ok(())
}

fn run_app<B: Backend>(terminal: &mut Terminal<B>, mut app: Tree) -> io::Result<()> {
    loop {
        terminal.draw(|f| app.render(f).unwrap_or(()))?;

        let show_hidden = app.show_hidden;
        if let Event::Key(key) = event::read()? {
            match app.mode {
                Mode::NORMAL => match key.code {
                    // modes
                    KeyCode::Esc | KeyCode::Char('q') => return Ok(()),
                    KeyCode::Char('/') => {
                        app.mode = Mode::SEARCH;
                        app.cwd().state.select(Some(0))
                    }
                    KeyCode::Char('H') => app.toggle_show_hidden(),
                    // motion
                    KeyCode::Left | KeyCode::Char('h') => app.cd_parent(),
                    KeyCode::Down | KeyCode::Char('j') => app.cwd().next(show_hidden),
                    KeyCode::Up | KeyCode::Char('k') => app.cwd().previous(show_hidden),
                    // TODO: handle symlinks
                    KeyCode::Right | KeyCode::Char('l') => app.cd_selected(),
                    _ => {}
                },
                Mode::SEARCH => match key.code {
                    KeyCode::Char(c) => {
                        if key.modifiers == KeyModifiers::CONTROL {
                            match c {
                                'n' => app.cwd().next(show_hidden),
                                'p' => app.cwd().previous(show_hidden),
                                'c' => {
                                    app.query = String::new();
                                    app.mode = Mode::NORMAL;
                                }
                                _ => {}
                            }
                        } else {
                            app.query.push(c)
                        }
                    }
                    // this should cd into directories and open files in EDITOR
                    // KeyCode::Enter => {app.cwd.files}
                    KeyCode::Backspace => {
                        app.query.pop();
                    }
                    KeyCode::Esc => {
                        app.query = String::new();
                        app.mode = Mode::NORMAL;
                    }
                    _ => {}
                },
            }
        }
    }
}
