use fm::FM;

use anyhow::Result;

mod finder;
mod fm;
mod log;

fn main() -> Result<()> {
    FM::new()?.run()?;
    Ok(())
}
