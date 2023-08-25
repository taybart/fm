use std::fs::File;
use std::io::prelude::*;

fn open() -> Result<File, String> {
    let file = File::options()
        .append(true)
        .open("/Users/taylor/dev/taybart/fm/fm.log")
        .map_err(|e| format!("could not open log file {e}"))?;
    Ok(file)
}
cfg_if::cfg_if! {
if #[cfg(debug_assertions)] {
    pub fn error(msg: String) {
        let mut file = open().expect("open file");
        file.write_all("[ERROR] ".as_bytes()).expect("write failed");
        file.write_all(msg.as_bytes()).expect("write failed");
        file.write_all("\n".as_bytes()).expect("write failed");
    }
    pub fn write(msg: String) {
        let mut file = open().expect("open file");
        file.write_all(msg.as_bytes()).expect("write failed");
        file.write_all("\n".as_bytes()).expect("write failed");
    }
} else {
    pub fn error(msg: String) {}
    pub fn write(msg: String) {}
}}
