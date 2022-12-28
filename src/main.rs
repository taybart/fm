use fs::tree::Tree;

use std::error::Error;

mod finder;
mod fs;
mod log;

fn main() -> Result<(), Box<dyn Error>> {
    let mut app = Tree::new();

    let res = app.run();

    if let Err(err) = res {
        println!("{:?}", err)
    }
    Ok(())
}
