// Binary target so `cargo install --path member-a` works — this is how the workspace
// multi-module build-info test drives a deps-collecting (install) command.
fn main() {
    println!("cli-cargo-member-a");
}
