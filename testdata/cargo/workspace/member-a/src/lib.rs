//! Workspace member A — depends on serde_json (external) and sibling member-b (path).

pub fn member_a() -> String {
    serde_json::json!({ "member": "a", "sibling": cli_cargo_member_b::member_b() }).to_string()
}
