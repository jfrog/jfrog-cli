#!/bin/bash

log(){
	echo "$1"
}

# Use the given key to configure the rpm macro. This is needed to sign an rpm.
# Arguments:
#   - gpgKeyFile   : key file location (in PEM format) to be used for signing the rpm
#                    The structure of the key content should be as follows,
#                        -----BEGIN PGP PUBLIC KEY BLOCK-----
#                        Version: GnuPG v1.4.7 (MingW32)
#                        .....
#                        -----END PGP PUBLIC KEY BLOCK-----
#                        -----BEGIN PGP PRIVATE KEY BLOCK-----
#                        Version: GnuPG v1.4.7 (MingW32)
#                        .....
#                        -----END PGP PRIVATE KEY BLOCK-----
#   - keyID : id of the provided key
rpmInitSigning(){
    local gpgKeyFile="${KEY_FILE}"
    local keyID="${KEY_ID}"

    log "Initializing rpm sign..."

    gpg --allow-secret-key-import --import ${gpgKeyFile} && \
    gpg --export -a "${keyID}" > /tmp/tmpFile && \
    rpm --import /tmp/tmpFile && \
    rpm -q gpg-pubkey --qf '%{name}-%{version}-%{release} --> %{summary}\n' | grep "${keyID}" || \
      { echo "ERROR: RPM signature initialization failed!" >&2; exit 1; }

    rpmEditRpmMacro "${keyID}" || \
      { echo "ERROR: Configuring rpm macro failed!" >&2; exit 1; }
}

rpmEditRpmMacro(){
    local keyID="$1"
    log "Configuring rpm macro for rpm sign"
    tee -a ~/.rpmmacros << RPM_MACRO_CONTENT
%_signature gpg
%_gpg_path /root/.gnupg
%_gpg_name ${keyID}
%_gpgbin /usr/bin/gpg
RPM_MACRO_CONTENT
}

expect_script() {
    cat << End-of-text #No white space between << and End-of-text
spawn rpm --resign $RPM_FILE_SIGNED
expect -exact "Enter pass phrase: "
send -- "$PASSPHRASE\r"
expect eof
exit
End-of-text

}

sign_rpm() {
    echo "Signing RPM..."
    cp -f "${RPM_FILE}" "${RPM_FILE_SIGNED}" || \
          { echo "ERROR: Copying ${RPM_FILE} to ${RPM_FILE_SIGNED} failed! " >&2; exit 1; }
    expect_script | /usr/bin/expect -f -
    cp -f "${RPM_FILE_SIGNED}" "${RPM_FILE}" || \
          { echo "ERROR: Copying ${RPM_FILE_SIGNED} to ${RPM_FILE} failed! " >&2; exit 1; }
}

KEY_FILE="${1}"
KEY_ID="${2}"
export PASSPHRASE="${3}"
RPM_FILE="${4}"
RPM_FILE_SIGNED="/tmp/jfrog-cli-rpm-signed.rpm"
rpmInitSigning
sign_rpm
