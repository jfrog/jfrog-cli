Name:           jfrog-cli-v2
Version:        %{cli_version}
Release:        %{cli_release}
Summary:        Smart client that provides a simple interface that automates access to JFrog products
Vendor:         JFrog Ltd.
Group:          Development/Tools
License:        Apache-2.0 License
URL:            http://www.jfrog.org
BuildRoot:      %{_tmppath}/build-%{name}-%{version}
BuildArch:      %{build_arch}

%define source_cli %{_sourcedir}/jfrog
%define _rpmfilename %{filename_prefix}-%{full_version}.rpm

%define target_cli_bin_dir /usr/bin
%define target_cli_bin %{target_cli_bin_dir}/jfrog

%description
JFrog CLI is a compact and smart client that provides a simple interface that automates access to JFrog products simplifying your automation scripts and making them more readable and easier to maintain.

%install
%__rm -rf %{buildroot}
%__install -d "%{buildroot}%{target_cli_bin_dir}"
%__cp -f %{source_cli} "%{buildroot}%{target_cli_bin}"

%posttrans

echo post transaction %{name} \$1 = $1 >>/tmp/rpminst

exit 0

%triggerpostun -- jfrog-cli-v2
echo trigger post uninstall %{name} \$1 = $1 >>/tmp/rpminst
exit 0

%triggerun -- jfrog-cli-v2
echo trigger uninstall %{name} \$1 = $1 >>/tmp/rpminst
exit 0

%files
%attr(755,root,root) %{target_cli_bin}

%doc
