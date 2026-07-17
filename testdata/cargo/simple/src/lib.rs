//! Minimal dependency-free library crate used as a Cargo integration-test fixture.

/// Adds two numbers. Exists only so the crate compiles and `cargo package` produces a .crate.
pub fn add(left: u64, right: u64) -> u64 {
    left + right
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_adds() {
        assert_eq!(add(2, 2), 4);
    }
}
