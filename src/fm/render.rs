use crossterm::{
    event::{DisableMouseCapture, EnableMouseCapture},
    execute,
    terminal::{disable_raw_mode, enable_raw_mode, EnterAlternateScreen, LeaveAlternateScreen},
};
use std::io::{self, Stdout};
use tui::{
    backend::{Backend, CrosstermBackend},
    layout::{Constraint, Direction, Layout, Rect},
    style::{Color, Modifier, Style},
    text::Text,
    widgets::Paragraph,
    Frame, Terminal,
};

use anyhow::Result;

use super::{dir::Dir, fm::FM, state::Mode, tree::Tree};

pub fn setup_tui() -> io::Result<Terminal<CrosstermBackend<Stdout>>> {
    enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen, EnableMouseCapture)?;
    let backend = CrosstermBackend::new(stdout);
    let terminal = Terminal::new(backend)?;
    Ok(terminal)
}
pub fn teardown_tui(terminal: &mut Terminal<CrosstermBackend<Stdout>>) -> io::Result<()> {
    disable_raw_mode()?;
    execute!(
        terminal.backend_mut(),
        LeaveAlternateScreen,
        DisableMouseCapture
    )?;
    terminal.show_cursor()?;
    Ok(())
}

pub fn render<B: Backend>(fm: &mut FM, f: &mut Frame<B>) -> Result<()> {
    let hide_parent = fm.state.hide_parent;
    // TODO: add to config
    if f.size().width < 80 {
        fm.state.hide_parent = true;
    }

    let constraints = if fm.state.hide_parent {
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

    let show_hidden = fm.state.show_hidden;
    /*
     * Left column
     */
    if !fm.state.hide_parent {
        fm.tree.parent().render(f, false, chunks[0], show_hidden);
    }

    /*
     * middle column
     */
    let idx = if fm.state.hide_parent { 0 } else { 1 };
    match fm.state.mode {
        Mode::Search => {
            fm.tree
                .cwd()
                .render_with_query(f, &fm.state.query_string, chunks[idx], show_hidden);
            f.render_widget(
                Paragraph::new(Text::raw(format!("/{}", &fm.state.query_string)))
                    .style(Style::default().add_modifier(Modifier::UNDERLINED)),
                Rect::new(chunks[idx].x, f.size().height - 1, chunks[idx].width, 1),
            );
            f.set_cursor(
                chunks[idx].x + fm.state.query_string.len() as u16 + 1,
                f.size().height,
            )
        }
        Mode::Normal => {
            fm.tree.cwd().render(f, true, chunks[idx], show_hidden);
        }
        Mode::Command => {
            fm.tree.cwd().render(f, true, chunks[idx], show_hidden);
            f.render_widget(
                Paragraph::new(Text::raw(format!(":{}", &fm.cmd.string)))
                    .style(Style::default().add_modifier(Modifier::UNDERLINED)),
                Rect::new(chunks[0].x, f.size().height - 1, chunks[idx].width, 1),
            );
            f.set_cursor(
                chunks[0].x + fm.cmd.string.len() as u16 + 1,
                f.size().height,
            )
        }
    };

    /*
     * right column
     */

    let idx = if fm.state.hide_parent { 1 } else { 2 };
    if let Some(selected) = fm
        .tree
        .cwd()
        .get_selected_file(show_hidden, &fm.state.query_string)
    {
        if selected.is_dir {
            Dir::new(selected.path)?.render(f, false, chunks[idx], show_hidden);
        } else {
            f.render_widget(Paragraph::new(Text::raw(selected.contents())), chunks[idx]);
        }
    };

    fm.state.hide_parent = hide_parent;
    Ok(())
}
