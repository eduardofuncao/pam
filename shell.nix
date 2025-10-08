{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    postgresql_15
    go
  ];

  # Disable _FORTIFY_SOURCE for debugging
  hardeningDisable = [ "fortify" ];

  shellHook = ''
    export PGDATA=$PWD/postgres_data
    export PGHOST=$PWD/postgres
    export LOG_PATH=$PWD/postgres/LOG
    export PGDATABASE=postgres
    export DATABASE_URL="postgresql:///sampledb?host=$PGHOST"

    if [ ! -d $PGHOST ]; then
      mkdir -p $PGHOST
    fi

    if [ ! -d $PGDATA ]; then
      echo "Initializing PostgreSQL database..."
      initdb $PGDATA --auth=trust >/dev/null
    fi

    pg_ctl start -l $LOG_PATH -o "-c unix_socket_directories=$PGHOST -c listen_addresses= -c port=5432"
    
    echo "PostgreSQL started successfully!"
    
    # Check if sampledb exists, if not run init.sql
    if ! psql -lqt | cut -d \| -f 1 | grep -qw sampledb; then
      echo ""
      echo "Setting up sample database with initial data..."
      psql -f init.sql
    else
      echo "Sample database already exists."
    fi
    
    echo ""
    echo "========================================="
    echo "PostgreSQL is ready!"
    echo "========================================="
    echo "Database URL: $DATABASE_URL"
    echo ""
    echo "Useful commands:"
    echo "  psql sampledb        - Connect to sample database"
    echo "  psql                 - Connect to default postgres database"
    echo "  pg_ctl stop          - Stop PostgreSQL server"
    echo ""
    echo "Sample queries to try:"
    echo "  SELECT * FROM users;"
    echo "  SELECT u.username, p.title FROM users u JOIN posts p ON u.id = p.user_id;"
    echo ""
  '';

  exitHook = ''
    pg_ctl stop
    echo "PostgreSQL stopped"
  '';
}
