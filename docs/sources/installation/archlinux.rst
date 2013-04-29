.. _arch_linux:

Arch Linux
==========

  Please note this is a community contributed installation path. The only 'official' installation is using the
  :ref:`ubuntu_linux` installation path. This version may sometimes be out of date.


Installing on Arch Linux is not officially supported but can be handled via 
either of the following AUR packages:

* `lxc-docker <https://aur.archlinux.org/packages/lxc-docker/>`_
* `lxc-docker-git <https://aur.archlinux.org/packages/lxc-docker-git/>`_

The lxc-docker package will install the latest tagged version of docker. 
The lxc-docker-git package will build from the current master branch.

Dependencies
------------

Docker depends on several packages which are specified as dependencies in
either AUR package.

* aufs3
* bridge-utils
* go
* iproute2
* linux-aufs_friendly
* lxc

Installation
------------

The instructions here assume **yaourt** is installed.  See 
`Arch User Repository <https://wiki.archlinux.org/index.php/Arch_User_Repository#Installing_packages>`_
for information on building and installing packages from the AUR if you have not
done so before.

Keep in mind that if **linux-aufs_friendly** is not already installed that a
new kernel will be compiled and this can take quite a while.

::

    yaourt -S lxc-docker-git


Starting Docker
---------------

Prior to starting docker modify your bootloader to use the 
**linux-aufs_friendly** kernel and reboot your system.

There is a systemd service unit created for docker.  To start the docker service:

::

    sudo systemctl start docker


To start on system boot:

::

    sudo systemctl enable docker
