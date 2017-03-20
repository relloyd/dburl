// Package dburl provides a standard, URL style mechanism for parsing and
// opening SQL database connection strings.
//
// Database URL Connection Strings
//
// Supported database URLs are of the form:
//
//   protocol+transport://user:pass@host/dbname?opt1=a&opt2=b
//   protocol:/path/to/file
//
// Where:
//
//   protocol  - driver name or alias (see below)
//   transport - the transport protocol [tcp, udp, unix] (only mysql for now)
//   user      - the username to connect as
//   pass      - the password to use
//   host      - the remote host
//   dbname*    - the database, instance, or service name/id to connect to
//   ?opt1=... - additional database driver options
//                 (see respective SQL driver for available options)
//
// * for Microsoft SQL Server, the syntax to supply an instance and database
// name is /instance/dbname, where /instance is optional. For Oracle databases,
// /dbname is the unique database ID (SID). Please see below for examples.
//
// Quickstart
//
// URLs in the above format can be parsed with Parse as such:
//
//   u, err := dburl.Parse("postgresql://user:pass@localhost/mydatabase/?sslmode=disable")
//   if err != nil { /* ... */ }
//
// Additionally, a simple helper func, Open, is available to simply parse,
// open, and return the SQL database connection:
//
//   db, err := dburl.Open("sqlite:mydatabase.sqlite3?loc=auto")
//   if err != nil { /* ... */ }
//
// Example URLs
//
// The following are URLs that can be handled with a call to Open or Parse:
//
//   postgres://user:pass@localhost/dbname
//   pg://user:pass@localhost/dbname?sslmode=disable
//   mysql://user:pass@localhost/dbname
//   mysql:/var/run/mysqld/mysqld.sock
//   sqlserver://user:pass@remote-host.com/dbname
//   mssql://user:pass@remote-host.com/instance/dbname
//   ms://user:pass@remote-host.com:port/instance/dbname?keepAlive=10
//   oracle://user:pass@somehost.com/oracledb
//   sap://user:pass@localhost/dbname
//   sqlite:/path/to/file.db
//   file:myfile.sqlite3?loc=auto
//   odbc+postgres://user:pass@localhost:port/dbname?option1=
//
// Protocol Schemes and Aliases
//
// The following protocols schemes (ie, driver) and their associated aliases
// are supported out of the box:
//
//   Database (driver)            | Protocol (scheme) Aliases
//   -----------------------------|------------------------------------
//   Microsoft SQL Server (mssql) | ms, sqlserver
//   MySQL (mysql)                | my, mariadb, maria, percona, aurora
//   Oracle (ora)                 | or, oracle, oci8, oci
//   PostgreSQL (postgres)        | pg, postgresql, pgsql
//   SQLite3 (sqlite3)            | sq, sqlite, file
//   -----------------------------|------------------------------------
//   Google Spanner (spanner)     | gs, google, span (not yet public)
//   -----------------------------|------------------------------------
//   Apache Avatica (avatica)     | av, phoenix
//   ClickHouse (clickhouse)      | ch
//   CockroachDB (cockroachdb)    | cr, cockroach, crdb, cdb
//   Couchbase (n1ql)             | n1, couchbase
//   Firebird SQL (firebirdsql)   | fb, firebird
//   MemSQL (memsql)              | me
//   Microsoft ADODB (adodb)      | ad, ado
//   ODBC (odbc)                  | od
//   OLE ODBC (oleodbc)*          | oo, ole, oleodbc
//   Cznic QL (ql)                | ql
//   SAP HANA (hdb)               | sa, saphana, sap, hana
//   Sybase SQL Anywhere (sqlany) | sy, sybase, any
//   VoltDB (voltdb)              | vo, volt, vdb
//   YQL (yql)                    | yq
//
// Any protocol scheme alias:// can be used in place of protocol://, and will
// work identically with Parse/Open.
//
// Usage
//
// Please note that the dburl package does not import actual SQL drivers, and
// only provides a standard way to parse/open respective database connection URLs.
//
// For reference, these are the following "expected" SQL drivers that would need
// to be imported:
//
//   Database (driver)            | Package
//   -----------------------------|-------------------------------------------------
//   Microsoft SQL Server (mssql) | github.com/denisenkom/go-mssqldb
//   MySQL (mysql)                | github.com/go-sql-driver/mysql
//   Oracle (ora)                 | gopkg.in/rana/ora.v4
//   PostgreSQL (postgres)        | github.com/lib/pq
//   SQLite3 (sqlite3)            | github.com/mattn/go-sqlite3
//   -----------------------------|-------------------------------------------------
//   Google Spanner (spanner)     | github.com/knq/spanner (not yet public)
//   -----------------------------|-------------------------------------------------
//   Apache Avatica (avatica)     | github.com/Boostport/avatica
//   ClickHouse (clickhouse)      | github.com/kshvakov/clickhouse
//   CockroachDB (cockroachdb)    | github.com/lib/pq
//   Couchbase (n1ql)             | github.com/couchbase/go_n1ql
//   Firebird SQL (firebirdsql)   | github.com/nakagami/firebirdsql
//   MemSQL (memsql)              | github.com/go-sql-driver/mysql
//   Microsoft ADODB (adodb)      | github.com/mattn/go-adodb
//   ODBC (odbc)                  | github.com/alexbrainman/odbc
//   OLE ODBC (oleodbc)*          | github.com/mattn/go-adodb
//   Cznic QL (ql)                | github.com/cznic/ql
//   SAP HANA (hdb)               | github.com/SAP/go-hdb/driver
//   Sybase SQL Anywhere (sqlany) | github.com/a-palchikov/sqlago
//   VoltDB (voltdb)              | github.com/VoltDB/voltdb-client-go/voltdbclient
//   YQL (yql)                    | github.com/mattn/go-yql
//
// * OLE ODBC is not an actual protocol, but instead is an alias for using the
//   "MSDASQL.1" OLE provider with the ADODB driver, and the DSN will be an
//   ADODB DSN, but with "Extended Properties" for the respective ODBC
//   parameters, including the underlying transport prootocol. As such,
//   "oleodbc+protocol://user:pass@host/dbname" URLs are equivalent to
//   "adodb://MSDASQL.1/?Extended+Properties=...". Please see the
//   documentation for GenOLEODBC for information regarding how URL components
//   are mapped/passed to ADODB's Extended Properties parameter.
//
// URL Parsing Rules
//
// Parse and Open rely heavily on the standard net/url.URL type, as such
// parsing rules have the same conventions/semantics as any URL parsed by the
// standard library's net/url.Parse.
//
// Related Projects
//
// This package was written mainly to support xo (https://github.com/knq/xo)
// and usql (https://github.com/knq/usql).
package dburl

import (
	"database/sql"
	"errors"
)

var (
	// ErrInvalidDatabaseScheme is the invalid database scheme error.
	ErrInvalidDatabaseScheme = errors.New("invalid database scheme")

	// ErrInvalidTransportProtocol is the invalid transport protocol error.
	ErrInvalidTransportProtocol = errors.New("invalid transport protocol")

	// ErrUnknownDatabaseScheme is the unknown database type error.
	ErrUnknownDatabaseScheme = errors.New("unknown database scheme")

	// ErrInvalidPort is the invalid port error.
	ErrInvalidPort = errors.New("invalid port")
)

// Open takes a urlstr like "protocol+transport://user:pass@host/dbname?option1=a&option2=b"
// and creates a standard sql.DB connection.
//
// See Parse for information on formatting URLs to work properly with Open.
func Open(urlstr string) (*sql.DB, error) {
	u, err := Parse(urlstr)
	if err != nil {
		return nil, err
	}

	return sql.Open(u.Driver, u.DSN)
}
