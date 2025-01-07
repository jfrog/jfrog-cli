#!/bin/bash

log(){
    echo "$1"
}

debug_info(){
    echo "=== DEBUG INFO ==="
    echo "Current User: $(whoami)"
    echo "GPG Version: $(gpg --version)"
    echo "GPG_TTY: $GPG_TTY"
    echo "TTY: $(tty)"
    echo "Files in /root/.gnupg:"
    ls -la /root/.gnupg
    echo "Environment Variables:"
    env
    echo "==================="
}

rpmInitSigning(){
    local gpgKeyFile="${KEY_FILE}"
    local keyID="${KEY_ID}"

    log "Initializing rpm sign..."

    # Start the GPG agent
    local gpg_agent_output
    gpg_agent_output=$(gpg-agent --daemon --allow-preset-passphrase)
    eval "$gpg_agent_output"

    # Set GPG_TTY if possible
    local tty_value
    if tty -s; then
        tty_value=$(tty)
        export GPG_TTY="$tty_value"
    else
        export GPG_TTY="/dev/null"
    fi

    # Debug info
    debug_info

    # Import the GPG key
    gpg --batch --import "${gpgKeyFile}" || { echo "ERROR: Failed to import GPG key"; exit 1; }
    gpg --batch --export -a "${keyID}" > /tmp/tmpFile || { echo "ERROR: Failed to export GPG key"; exit 1; }
    if rpm --import /tmp/tmpFile && rpm -q gpg-pubkey --qf '%{name}-%{version}-%{release} --> %{summary}\n' | grep "${keyID}"; then
        echo "RPM signature initialization succeeded."
    else
        echo "ERROR: RPM signature initialization failed!" >&2
        exit 1
    fi

    rpmEditRpmMacro "${keyID}" || { echo "ERROR: Configuring rpm macro failed!" >&2; exit 1; }
}

rpmEditRpmMacro(){
    local keyID="$1"
    log "Configuring rpm macro for rpm sign"
    tee -a ~/.rpmmacros << RPM_MACRO_CONTENT
%_signature gpg
%_gpg_path /root/.gnupg
%_gpg_name ${keyID}
%_gpgbin /usr/bin/gpg
%_gpg_sign_cmd %{__gpg} gpg --batch --pinentry-mode loopback --passphrase-file /tmp/passphrase --detach-sign --armor --yes --no-secmem-warning -u %{_gpg_name} -o %{__signature_filename} %{__plaintext_filename}
RPM_MACRO_CONTENT
}

sign_rpm() {
    echo "Signing RPM..."
    echo "${PASSPHRASE}" > /tmp/passphrase
    cp -f "${RPM_FILE}" "${RPM_FILE_SIGNED}" || { echo "ERROR: Copying ${RPM_FILE} to ${RPM_FILE_SIGNED} failed! " >&2; exit 1; }
    
    gpg --batch --pinentry-mode loopback --passphrase-file /tmp/passphrase --detach-sign --armor --yes --no-secmem-warning -u "${KEY_ID}" -o "${RPM_FILE_SIGNED}.asc" "${RPM_FILE_SIGNED}" || { echo "ERROR: GPG signing failed!"; exit 1; }
    
    rpm --addsign "${RPM_FILE_SIGNED}" || { echo "ERROR: RPM signing failed!"; exit 1; }
    
    cp -f "${RPM_FILE_SIGNED}" "${RPM_FILE}" || { echo "ERROR: Copying ${RPM_FILE_SIGNED} to ${RPM_FILE} failed! " >&2; exit 1; }
    rm /tmp/passphrase
}

KEY_FILE="${1}"
KEY_ID="${2}"
export PASSPHRASE="${3}"
RPM_FILE="${4}"
RPM_FILE_SIGNED="/tmp/jfrog-cli-rpm-signed.rpm"

rpmInitSigning
sign_rpm
