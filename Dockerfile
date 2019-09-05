FROM debian:jessie
ADD . /
CMD ["/usr/bin/mosoly-ledger-bridge"]
