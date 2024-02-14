Name:          bext 
Vendor:        ublue-os
Version:       {{{ bext_version }}}
Release:       0%{?dist}
Summary:       Manager for systemd system extensions 
License:       Apache-2.0
URL:           https://github.com/%{vendor}/%{name}
# Detailed information about the source Git repository and the source commit
# for the created rpm package
VCS:           {{{ git_dir_vcs }}}

# git_dir_pack macro places the repository content (the source files) into a tarball
# and returns its filename. The tarball will be used to build the rpm.
Source:        {{{ git_dir_pack }}}

BuildArch:     x86_64 
Supplements:   podman 
BuildRequires: systemd-rpm-macros
BuildRequires: btrfs-progs-devel 
BuildRequires: gpgme-devel
BuildRequires: device-mapper-devel 
BuildRequires: go 

%description
Installs and configures bext binary and systemd services

%global debug_package %{nil}

%prep
{{{ git_dir_setup_macro }}}

%build
go build -o %{name}

%install
install -D -m 0755 %{name} %{buildroot}%{_bindir}/%{name}
install -D -m 0644 service/%{name}-mount.service %{buildroot}%{_unitdir}/%{name}-mount.service

%files
%{_bindir}/%{name}
%{_unitdir}/%{name}-mount.service
%attr(0755,root,root) %{_bindir}/%{NAME}
%attr(0644,root,root) %{_exec_prefix}/lib/systemd/system/%{NAME}-mount.service

%changelog
%autochangelog
