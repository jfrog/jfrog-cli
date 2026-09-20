// Presence of a build script pulls [build-dependencies] (cc) into the graph with
// scope "build", so the integration tests can verify build-scope handling.
fn main() {
    println!("cargo:rerun-if-changed=build.rs");
}
