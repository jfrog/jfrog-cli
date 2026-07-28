// Binary target so `cargo install --path .` works (the deps-collection path in tests).
fn main() {
    let v = serde_json::json!({ "cli-cargo-lib": true });
    println!("{}", v);
}
