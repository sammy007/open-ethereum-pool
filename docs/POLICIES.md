# Enforcing Policies

Pool policy server collecting several stats on per IP basis. There are two options: `iptables+ipset` or simple application level bans. Banning is disabled by default.

## Firewall Banning

First you need to configure your firewall to use `ipset`, read [this article](https://wiki.archlinux.org/index.php/Ipset).

Specify `ipset` name for banning in `policy` section. Timeout argument (in seconds) will be passed to this `ipset`. Stratum will use `os/exec` command like `sudo ipset add banlist x.x.x.x 1800` for banning, so you have to configure `sudo` properly and make sure that your system will never ask for password:

Example `/etc/sudoers.d/pool` where `pool` is a username under which pool runs:

    pool ALL=NOPASSWD: /sbin/ipset

If you need something simple, just set `ipset` name to blank string and simple application level banning will be used instead.

## Limiting

Under some weird circumstances you can enforce limits to prevent connection flood to stratum, there are initial settings: `limit` and `limitJump`. Policy server will increase number of allowed connections per IP address on each valid share submission. Stratum will not enforce this policy for a `grace` period specified after stratum start.
