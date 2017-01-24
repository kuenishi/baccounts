# Baccounts

Internet account (password) manager, with following goals:

* Secure: Manage secure passwords (secure = long and complex enough)
* **SMALL**: as minimal code required; ~ 1000 LOC for core functionalities
* Dependent on a few small libraries as less as possible
* Platform independent: work on Darwin, Windows, Linux or Unix
* Portable and readable, not breakable persistent data

Under following assumptions or limitations

* Trust Go standard library
* No GUI
* Only work with GnuPG 2.0 / 1.4 generated keyring files (PGP format; GnuPG 2.1 uses .kbx gpgsm format)
* [It's a hustle working on gpgsm format](https://github.com/kubernetes/helm/issues/1592)
* [on gpg-agent error of pinentry](https://wiki.archlinuxjp.org/index.php/GnuPG#gpg-agent)

# Build and install

```sh
$ git clone git://github.com/kuenishi/baccounts
$ cd baccounts
$ go get
$ go build
$ go install
```

# Related Products

* [passgo](https://github.com/ejcx/passgo)
* [pass](https://www.passwordstore.org/)
* KeePass
* 1Password

# LICENSE

GPL version 3

# TODO

* update password (create a new one)
* Export to other devices that does not have secret keys (Android, other computers)
* how to share between devices like Android phone?

# Test keys

* kuenishi@example.com - baccounts
* kuenishi@example.com - baccounts
* test@example.com - test2
