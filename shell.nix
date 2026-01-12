let
  pkgs = import <nixpkgs> { config.allowUnfree = true; };
  oracleLib = pkgs.oracle-instantclient.lib;
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    oracle-instantclient
    postgresql_15
    go
    sqlite
    duckdb
    arrow-cpp 
  ];

  hardeningDisable = [ "fortify" ];

  shellHook = ''
    export LD_LIBRARY_PATH=${oracleLib}/lib:$LD_LIBRARY_PATH
    export ORACLE_HOME=${oracleLib}

    echo "========================================="
    echo "Pam development environment ready!"
    echo "========================================="
    echo ""
    echo "Available tools:"
    echo "  - Go compiler"
    echo "  - PostgreSQL client (psql)"
    echo "  - SQLite client (sqlite3)"
    echo "  - Oracle Instant Client"
    echo ""
    echo "To test Pam with real databases, use dbeesly:"
    echo "  https://github.com/dbeesly"
    echo ""
  '';
}
