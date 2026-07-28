//! Workspace member B — depends on serde_json (external, from Artifactory).

pub fn member_b() -> String {
    serde_json::json!({ "member": "b" }).to_string()
}
