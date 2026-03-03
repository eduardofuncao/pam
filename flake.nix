{
  description = "Squix's SQL Stash - SQL query CLI tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;  # For Oracle Instant Client
        };
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "squix";
          version = "0.3.0-beta";

          src = ./.;

          # Run: nix build .#default 2>&1 | grep "got:" to get real hash
          vendorHash = "sha256-CRvFH9Fn7iGZck6DBTh+2yNRj6J/qqLs6Z1dgfePUYs=";


          # Native dependencies
          buildInputs = with pkgs; [
            sqlite.dev        # For go-sqlite3
            duckdb           # For go-duckdb
            arrow-cpp        # DuckDB dependency
            oracle-instantclient.lib  # For godror
          ];

          # Linker flags
          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${self.packages.${system}.default.version}"
          ];

          # Oracle library paths
          propagatedBuildInputs = with pkgs; [ oracle-instantclient.lib ];

          meta = with pkgs.lib; {
            description = "Minimal CLI tool for managing SQL queries across multiple databases";
            homepage = "https://github.com/eduardofuncao/squix";
            license = licenses.mit;
            mainProgram = "squix";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            sqlite.dev
            duckdb
            arrow-cpp
            oracle-instantclient.lib
            postgresql
          ];

          shellHook = ''
            export LD_LIBRARY_PATH=${pkgs.oracle-instantclient.lib}/lib:$LD_LIBRARY_PATH
            export ORACLE_HOME=${pkgs.oracle-instantclient.lib}
            export CGO_ENABLED=1

            echo "========================================="
            echo "Squix development environment ready!"
            echo "========================================="
            echo ""
            echo "Available tools:"
            echo "  - Go compiler"
            echo "  - PostgreSQL client (psql)"
            echo "  - SQLite client (sqlite3)"
            echo "  - Oracle Instant Client"
            echo ""
          '';
        };
      }
    );
}
