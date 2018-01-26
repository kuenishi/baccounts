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
* Only work with GnuPG 2.0 / 1.4 generated keyring files (PGP format; GnuPG 2.1 uses .kbx gpgsm format) -> Workaround below
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
* [Bitwarden](https://bitwarden.com/) whose [CLI version](https://github.com/bitwarden/cli) will override this product

# LICENSE

GPL version 3

# Workaround on later GnuPG key format

GnuPG >= 2.2 has a new public and secret key format instead of
`$HOME/.gnupg/pubring.gpg` and `$HOME/.gnupg/secring.gpg`, while
baccounts still reads secret keys from it (This is because Go openpgp
module only supports PGP compatible format). But GnuPG supports
exporting secret key to old format, like:

```sh
$ gpg --export > ~/.gnupg/pubring.gpg
$ gpg --export-secret-keys > ~/.gnupg/secring.gpg
```

- [export-secret-keys](https://www.gnupg.org/gph/en/manual/r887.html)
- [Removal of the secret keyring](https://www.gnupg.org/faq/whats-new-in-2.1.html#nosecring)
- [Secring does not exist anymore with the latest gnuPG version](https://github.com/jcmdev0/gpgagent/issues/2#issuecomment-306054405)

# TODO

* update password (create a new one)
* Export to other devices that does not have secret keys (Android, other computers)
* how to share between devices like Android phone?

# Test keys

* kuenishi@example.com - baccounts
* test@example.com - test2
