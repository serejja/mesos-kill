Mesos-kill
==========

Mesos-kill is a simple Go tool for killing Mesos frameworks. Basically it just wraps Mesos Master `teardown` call and allows killing frameworks not by their ids but names.

Installation
------------

Assuming your `$GOPATH/bin` is in your `$PATH` this should be enough:

```
# go get github.com/serejja/mesos-kill
```

Usage
-----

```
# mesos-kill [<master>] <framework-name-regex>

<master>: host:port pair for Mesos Master node. If not specified will check MESOS_MASTER env and fall back to 127.0.0.1:5050 if not set
<framework-name-regex>: name or regular expression of framework to kill. It is ok to match multiple frameworks
```

Examples
--------

You may use `mesos-kill` in 3 different ways:

1. Assuming Mesos Master runs on localhost you may just run `mesos-kill framework` and this will query Mesos Master at `127.0.0.1:5050` to kill the given framework.
2. You may also `export MESOS_MASTER=master:5050` and then run `mesos-kill framework`. This will query Mesos Master at `master:5050` instead of `127.0.0.1:5050`.
3. You may just use `mesos-kill master:5050 framework` to provide Mesos Master host:port pair directly.