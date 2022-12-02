mod consts;
pub mod matcher;
pub mod matrix;
pub mod scorer;

pub type MatchWithPositions<'a> = (f64, &'a str, Vec<usize>);

pub fn match_and_score_with_positions<'a>(
    needle: &str,
    haystack: &'a str,
) -> Option<MatchWithPositions<'a>> {
    if matcher::matches(needle, haystack) {
        let (score, positions) = scorer::score_with_positions(needle, haystack);
        Some((score, haystack, positions))
    } else {
        None
    }
}
