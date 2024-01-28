# Baccounts v2

Internet account (password) manager, with following goals:

* Secure: Manage secure passwords (secure = long and complex enough)
* **SMALL**: as minimal code required; ~ 1000 LOC for core functionalities
* Dependent on a few small libraries as less as possible
* Platform independent: work on Darwin, Windows, Linux or Unix
* Portable and readable, not breakable persistent data

Under following assumptions or limitations

* Trust Rust standard library
* No GUI
* [on gpg-agent error of pinentry](https://wiki.archlinuxjp.org/index.php/GnuPG#gpg-agent)

# Build and install

```sh
$ cargo install
```

# Changes from version 1

- Rewrittin in Rust
- Stop using GnuPG keyring and need just private & public key files
- Moved primary configuration directory to `$XDG_CONFIG_DIR/baccounts/`

# Related Products

* [passgo](https://github.com/ejcx/passgo)
* [pass](https://www.passwordstore.org/)
* KeePass
* 1Password
* [Bitwarden](https://bitwarden.com/) whose [CLI version](https://github.com/bitwarden/cli) will override this product

# LICENSE

GPL version 3
