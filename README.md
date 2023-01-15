# protoc-gen-go-ddl

> Protoc Plugin for generating DDL and Golang database models.

protoc-gen-go-ddl is a plugin of [ProtocolBuffer compiler](https://github.com/protocolbuffers/protobuf).

It generates protobuf messages into various datastores' DDL (Data Definition Language) and DML (Data Manipulation Language) code stubs.

## Prerequisites

- protoc

## Installation

```shell
go get github.com/taehoio/protoc-gen-go-ddl
```

## Usage

We recommend using [buf](https://github.com/bufbuild/buf).

### buf

```yaml
# in buf.gen.yaml
version: v1
plugins:
  - name: go
    out: gen/go
    opt:
      - paths=source_relative
  - name: go-ddl
    out: gen/go
    opt:
      - paths=source_relative
```

### CLI

```shell
protoc \
    --go-ddl_out=gen/go \
    --proto_path=proto \
    proto/taehoio/ddl/services/test/v1/test.proto
```

## Datastores

Currently, only `MySQL` is supported. We have a plan to support other datastores such as MongoDB, DynamoDB, Spanner, etc. for sure.

```protobuf
enum DatastoreType {
  DATASTORE_TYPE_UNSPECIFIED = 0;
  DATASTORE_TYPE_SQLITE = 1;
  DATASTORE_TYPE_MYSQL = 2;
  DATASTORE_TYPE_POSTGRESQL = 3;
  DATASTORE_TYPE_MSSQL = 4;
  DATASTORE_TYPE_ORACLE = 5;
  DATASTORE_TYPE_BIGQUERY = 6;
  DATASTORE_TYPE_MONGODB = 7;
}
```

## Data Types

The following protobuf types are supported:

- protobuf standard primitive types
  - uint32, uint64, int32, int64, float, double, bool, string
- protobuf optional types
  - optional string, etc.
- protobuf enum types
  - enum E {...}
  - E e = 1;
- [google timestamp type](https://github.com/golang/protobuf/blob/master/ptypes/timestamp/timestamp.proto)
  - google.protobuf.Timestamp
- [google date type](https://github.com/googleapis/googleapis/blob/master/google/type/date.proto)
  - google.type.Date

## Field Options

- `key`: Define a primary key of the table, `uint64` type only.
- `index`: Define an index of the table.
- `unique`: Define a unique index of the table.
  You can have `index` and `unique` together to define a unique index.

```protobuf
extend google.protobuf.FieldOptions {
  bool key = 61001;
  repeated string index = 61002;
  repeated string unique = 61003;
}
```

## Examples

```protobuf
syntax = "proto3";

package taehoio.ddl.services.test.v1;

import "taehoio/ddl/protobuf/v1/options.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";

option go_package = "github.com/taehoio/protoc-gen-go-ddl/gen/go/ddl/services/test/v1;testv1";

message User {
  option (taehoio.ddl.protobuf.v1.datastore_type) = DATASTORE_TYPE_MYSQL;

  int64 id = 1 [(taehoio.ddl.protobuf.v1.key) = true];

  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Timestamp updated_at = 3;
  google.protobuf.Timestamp deleted_at = 4;

  string password_hash = 5;
  string full_name = 6;
  string email = 7 [(taehoio.ddl.protobuf.v1.index) = "name=idx_email"];
}

enum CountryCode {
  COUNTRY_CODE_UNSPECIFIED = 0;
  COUNTRY_CODE_KR = 1;
  COUNTRY_CODE_CA = 2;
  COUNTRY_CODE_GB = 3;
  COUNTRY_CODE_JP = 4;
}

enum UserIdType {
  USER_ID_TYPE_UNSPECIFIED = 0;
  USER_ID_TYPE_KARROT = 1;
  USER_ID_TYPE_HOIAN = 2;
}

message UserCheckin {
  option (taehoio.ddl.protobuf.v1.datastore_type) = DATASTORE_TYPE_MYSQL;

  int64 id = 1 [(taehoio.ddl.protobuf.v1.key) = true];

  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Timestamp updated_at = 3;
  google.protobuf.Timestamp deleted_at = 4;

  CountryCode country_code = 5 [(taehoio.ddl.protobuf.v1.index) = "name=ix_countrycode_useridtype_userid_measuredat"];
  UserIdType user_id_type = 6 [(taehoio.ddl.protobuf.v1.index) = "name=ix_countrycode_useridtype_userid_measuredat"];
  string user_id = 7 [(taehoio.ddl.protobuf.v1.index) = "name=ix_countrycode_useridtype_userid_measuredat"];

  double latitude = 8;
  double longitude = 9;
  google.protobuf.DoubleValue altitude = 10;
  google.protobuf.DoubleValue horizontal_accuracy = 11;
  google.protobuf.DoubleValue vertical_accuracy = 12;
  google.protobuf.Timestamp measured_at = 13 [(taehoio.ddl.protobuf.v1.index) = "name=ix_countrycode_useridtype_userid_measuredat"];
}
```

It will generate codes below.

### DDL

```sql
CREATE TABLE `user` (
  `id` BIGINT NOT NULL,
  `created_at` TIMESTAMP(6) NULL DEFAULT NULL,
  `updated_at` TIMESTAMP(6) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP(6) NULL DEFAULT NULL,
  `password_hash` VARCHAR(255) NOT NULL,
  `full_name` VARCHAR(255) NOT NULL,
  `email` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`id`)
);

CREATE INDEX `idx_email` ON `user` (`email`);

CREATE TABLE `user_checkin` (
  `id` BIGINT NOT NULL,
  `created_at` TIMESTAMP(6) NULL DEFAULT NULL,
  `updated_at` TIMESTAMP(6) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP(6) NULL DEFAULT NULL,
  `country_code` INT NOT NULL,
  `user_id_type` INT NOT NULL,
  `user_id` VARCHAR(255) NOT NULL,
  `latitude` DOUBLE NOT NULL,
  `longitude` DOUBLE NOT NULL,
  `altitude` DOUBLE NULL DEFAULT NULL,
  `horizontal_accuracy` DOUBLE NULL DEFAULT NULL,
  `vertical_accuracy` DOUBLE NULL DEFAULT NULL,
  `measured_at` TIMESTAMP(6) NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE INDEX `ix_countrycode_useridtype_userid_measuredat` ON `user_checkin` (`country_code`, `user_id_type`, `user_id`, `measured_at`);
```

### Golang codes

```go

package testv1

import (
    "context"
    "database/sql"
    "strings"

    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type UserRecorder interface {
	Get(ctx context.Context, db *sql.DB, id int64) (*User, error)
	List(ctx context.Context, db *sql.DB, paginationOpts ...PaginationOption) ([]*User, error)
	FindByIDs(ctx context.Context, db *sql.DB, ids []int64) ([]*User, error)
	Save(ctx context.Context, db *sql.DB, message *User) error
	SaveTx(ctx context.Context, tx *sql.Tx, message *User) error
	Delete(ctx context.Context, db *sql.DB, id int64) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id int64) error
	FindOneByEmail(ctx context.Context, db *sql.DB, email interface{}) (*User, error)
	FindByEmail(ctx context.Context, db *sql.DB, email interface{}, paginationOpts ...PaginationOption) ([]*User, error)
	DeleteByEmail(ctx context.Context, db *sql.DB, email interface{}) error
}


type UserCheckinRecorder interface {
	Get(ctx context.Context, db *sql.DB, id int64) (*UserCheckin, error)
	List(ctx context.Context, db *sql.DB, paginationOpts ...PaginationOption) ([]*UserCheckin, error)
	FindByIDs(ctx context.Context, db *sql.DB, ids []int64) ([]*UserCheckin, error)
	Save(ctx context.Context, db *sql.DB, message *UserCheckin) error
	SaveTx(ctx context.Context, tx *sql.Tx, message *UserCheckin) error
	Delete(ctx context.Context, db *sql.DB, id int64) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id int64) error
	FindOneByCountryCodeAndUserIdTypeAndUserIdAndMeasuredAt(ctx context.Context, db *sql.DB, countryCode interface{}, userIdType interface{}, userId interface{}, measuredAtStartTime interface{}, measuredAtEndTime interface{}) (*UserCheckin, error)
	FindByCountryCodeAndUserIdTypeAndUserIdAndMeasuredAt(ctx context.Context, db *sql.DB, countryCode interface{}, userIdType interface{}, userId interface{}, measuredAtStartTime interface{}, measuredAtEndTime interface{}, paginationOpts ...PaginationOption) ([]*UserCheckin, error)
	DeleteByCountryCodeAndUserIdTypeAndUserIdAndMeasuredAt(ctx context.Context, db *sql.DB, countryCode interface{}, userIdType interface{}, userId interface{}, measuredAtStartTime interface{}, measuredAtEndTime interface{}) error
}

```

The entire codes are located in [testdata](/testdata) directory

## Caution

- You need to enable `parseTime` option when using mysql datastore.
  ```go
  db, err := sql.Open("mysql", "USER:PASSWORD@tcp(127.0.0.1:3306)/DATABASE?parseTime=true")
  ```

## TODOs

- [x] protobuf optional field support
- [ ] MongoDB support
- [ ] DynamoDB support

## Contributing

Any contribution will be welcomed!
