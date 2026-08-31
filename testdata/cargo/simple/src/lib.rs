//! Library crate used as a Cargo integration-test fixture. Depends on serde_json so
//! dependency-scope collection (prod/transitive) has something to capture.

/// Adds two numbers. Exists only so the crate compiles and `cargo package` produces a .crate.
pub fn add(left: u64, right: u64) -> u64 {
    left + right
}

/// Uses the serde_json production dependency so it is exercised at compile time.
pub fn tag() -> String {
    serde_json::json!({ "cli-cargo-lib": "lib" }).to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_adds() {
        assert_eq!(add(2, 2), 4);
    }
}
