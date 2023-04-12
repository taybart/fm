use fm::FM;

use std::error::Error;

mod finder;
mod fm;
mod log;

fn main() -> Result<(), Box<dyn Error>> {
    let mut fm = FM::new()?;

    let res = fm.run();

    if let Err(err) = res {
        println!("{:?}", err)
    }
    Ok(())
}
