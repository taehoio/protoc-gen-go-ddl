package ddl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	protobufv1 "github.com/taehoio/protoc-gen-go-ddl/gen/go/taehoio/ddl/protobuf/v1"
)

func ddlUserMessage() (*MessageInfo, error) {
	u, err := userMessage()
	if err != nil {
		return nil, err
	}
	return NewMessageInfo(*u)
}

func TestListMessageOptions(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	opts, err := listMessageOptions(*u)
	assert.NoError(t, err)
	assert.Len(t, opts, 1)
	assert.Equal(t, MessageOption{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}, opts[0])
}

func TestExtractFields(t *testing.T) {
	u, err := userMessage()
	assert.NoError(t, err)

	fields, err := extractFields(*u)
	assert.NoError(t, err)
	assert.Len(t, fields, 10)
}

func TestExtractKeyFields(t *testing.T) {
	u, err := ddlUserMessage()
	assert.NoError(t, err)

	fieldNames, err := extractKeyFields(u.Fields)
	assert.NoError(t, err)
	assert.Len(t, fieldNames, 1)
}

func TestExtractIndices(t *testing.T) {
	u, err := ddlUserMessage()
	assert.NoError(t, err)

	indices, err := extractIndices(u.Fields)
	assert.NoError(t, err)
	assert.Len(t, indices, 1)
}

func TestGenerateMySQLDDLSQL(t *testing.T) {
	expectedLines := []string{
		"DROP TABLE IF EXISTS `user`;",
		"",
		"CREATE TABLE `user` (",
		"	`id` BIGINT NOT NULL,",
		"	`created_at` TIMESTAMP(6) NOT NULL,",
		"	`updated_at` TIMESTAMP(6) NOT NULL,",
		"	`deleted_at` TIMESTAMP(6) NULL DEFAULT NULL,",
		"	`password_hash` VARCHAR(255) NOT NULL,",
		"	`full_name` VARCHAR(100) NULL DEFAULT NULL,",
		"	`email` VARCHAR(255) NOT NULL,",
		"	`profile_json` TEXT NOT NULL,",
		"	`birth_date` DATE NOT NULL,",
		"	`death_date` DATE NULL DEFAULT NULL,",
		"	PRIMARY KEY (`id`)",
		");",
		"",
		"CREATE INDEX `idx_email` ON `user` (`email`);",
	}
	expected := strings.Join(expectedLines, "\n")

	u, err := ddlUserMessage()
	assert.NoError(t, err)
	u.MessageOptions = []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}}

	s, err := u.generateMySQLDDL()
	assert.NoError(t, err)
	assert.Equal(t, expected, s)
}

func TestMessageInfo_supportsDatastore(t *testing.T) {
	type args struct {
		datastore protobufv1.DatastoreType
	}
	tests := []struct {
		name        string
		messageInfo MessageInfo
		args        args
		want        bool
	}{
		{
			name: "test mysql datastore",
			messageInfo: MessageInfo{
				MessageOptions: []MessageOption{
					{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"},
				},
			},
			args: args{
				datastore: protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL,
			},
			want: true,
		},
		{
			name: "unsupported datastore",
			messageInfo: MessageInfo{
				MessageOptions: []MessageOption{},
			},
			args: args{
				datastore: protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL,
			},
			want: false,
		},
		{
			name: "test postgres datastore",
			messageInfo: MessageInfo{
				MessageOptions: []MessageOption{
					{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_POSTGRESQL"},
				},
			},
			args: args{
				datastore: protobufv1.DatastoreType_DATASTORE_TYPE_POSTGRESQL,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := tt.messageInfo
			assert.Equalf(t, tt.want, mi.supportsDatastore(tt.args.datastore), "supportsDatastore(%v)", tt.args.datastore)
		})
	}
}

func TestMessageInfo_GenerateDDL(t *testing.T) {
	t.Run("fails due to unspecified datastore", func(t *testing.T) {
		u, err := ddlUserMessage()
		assert.NoError(t, err)
		u.MessageOptions = []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_UNSPECIFIED"}}

		got, err := u.GenerateDDL()
		assert.ErrorIs(t, err, ErrorDatastoreUnspecified, "GenerateDDL()")
		assert.Empty(t, got)
	})

	t.Run("fails due to unsupported datastore", func(t *testing.T) {
		u, err := ddlUserMessage()
		assert.NoError(t, err)
		u.MessageOptions = []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_SQLITE"}}

		got, err := u.GenerateDDL()
		assert.ErrorIs(t, err, ErrNotSupportedDatastore, "GenerateDDL()")
		assert.Empty(t, got)
	})

	t.Run("succeeded with mysql datastore", func(t *testing.T) {
		expectedLines := []string{
			"DROP TABLE IF EXISTS `user`;",
			"",
			"CREATE TABLE `user` (",
			"	`id` BIGINT NOT NULL,",
			"	`created_at` TIMESTAMP(6) NOT NULL,",
			"	`updated_at` TIMESTAMP(6) NOT NULL,",
			"	`deleted_at` TIMESTAMP(6) NULL DEFAULT NULL,",
			"	`password_hash` VARCHAR(255) NOT NULL,",
			"	`full_name` VARCHAR(100) NULL DEFAULT NULL,",
			"	`email` VARCHAR(255) NOT NULL,",
			"	`profile_json` TEXT NOT NULL,",
			"	`birth_date` DATE NOT NULL,",
			"	`death_date` DATE NULL DEFAULT NULL,",
			"	PRIMARY KEY (`id`)",
			");",
			"",
			"CREATE INDEX `idx_email` ON `user` (`email`);",
		}
		expected := strings.Join(expectedLines, "\n")

		u, err := ddlUserMessage()
		assert.NoError(t, err)
		u.MessageOptions = []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}}

		s, err := u.GenerateDDL()
		assert.NoError(t, err)
		assert.Equal(t, expected, s)
	})
}

func TestMessageInfo_getDatastoreOption(t *testing.T) {
	tests := []struct {
		name        string
		messageInfo MessageInfo
		want        protobufv1.DatastoreType
	}{
		{
			name: "message option with mysql",
			messageInfo: MessageInfo{
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}},
			},
			want: protobufv1.DatastoreType_DATASTORE_TYPE_MYSQL,
		},
		{
			name: "with empty message option",
			messageInfo: MessageInfo{
				MessageOptions: []MessageOption{},
			},
			want: protobufv1.DatastoreType_DATASTORE_TYPE_UNSPECIFIED,
		},
		{
			name: "with unknown data store",
			messageInfo: MessageInfo{
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_UNKNOWN"}},
			},
			want: protobufv1.DatastoreType_DATASTORE_TYPE_UNSPECIFIED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mi := tt.messageInfo
			assert.Equalf(t, tt.want, mi.getDatastoreOption(), "getDatastoreOption()")
		})
	}
}

func TestMessageInfo_DDLFileSuffix(t *testing.T) {
	tests := []struct {
		name    string
		mi      MessageInfo
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success with mysql",
			mi: MessageInfo{
				Name:           "user",
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}},
			},
			want: "_ddl_user_mysql.pb.sql",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.NoError(t, err, i...)
				return false
			},
		},
		{
			name: "error with unspecified",
			mi: MessageInfo{
				Name:           "user",
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_UNSPECIFIED"}},
			},
			want: "",
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				assert.ErrorIs(t, err, ErrNotSupportedDatastore, msgAndArgs...)
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mi.DDLFileSuffix()
			if !tt.wantErr(t, err, "DDLFileSuffix()") {
				return
			}
			assert.Equalf(t, tt.want, got, "DDLFileSuffix()")
		})
	}
}

func TestMessageInfo_generateMongodbDDL(t *testing.T) {
	expected := strings.Join([]string{
		`db.createCollection("user")`,
		`db.user.createIndex({"id":1})`,
		`db.user.createIndex({"email":1}, {"name":"idx_email"})`,
	}, "\n")

	u, err := ddlUserMessage()
	assert.NoError(t, err)
	u.MessageOptions = []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MONGODB"}}

	s, err := u.generateMongodbDDL()
	assert.NoError(t, err)
	assert.Equal(t, expected, s)
}

func TestMessageInfo_DMLFileSuffix(t *testing.T) {
	tests := []struct {
		name    string
		mi      MessageInfo
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success with mysql datastore",
			mi: MessageInfo{
				Name:           "User",
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}},
			},
			want: "_dml_user_mysql.pb.go",
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.NoError(t, err, msgAndArgs...)
			},
		},
		{
			name: "error with unspecified datastore",
			mi: MessageInfo{
				Name:           "User",
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_UNSPECIFIED"}},
			},
			want: "",
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrNotSupportedDatastore)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mi.DMLFileSuffix()
			if !tt.wantErr(t, err, "DMLFileSuffix()") {
				return
			}
			assert.Equalf(t, tt.want, got, "DMLFileSuffix()")
		})
	}
}

func TestMessageInfo_DMLMockFileSuffix(t *testing.T) {
	tests := []struct {
		name    string
		mi      MessageInfo
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success with mysql datastore",
			mi: MessageInfo{
				Name:           "User",
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_MYSQL"}},
			},
			want: "_dml_user_mysql_mock.pb.go",
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.NoError(t, err, msgAndArgs...)
			},
		},
		{
			name: "error with unspecified datastore",
			mi: MessageInfo{
				Name:           "User",
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_UNSPECIFIED"}},
			},
			want: "",
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrorDatastoreUnspecified, msgAndArgs...)
			},
		},
		{
			name: "error with not supported datastore",
			mi: MessageInfo{
				Name:           "User",
				MessageOptions: []MessageOption{{Name: "daangn.ddl.protobuf.v1.datastore_type", Value: "DATASTORE_TYPE_SQLITE"}},
			},
			want: "",
			wantErr: func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrNotSupportedDatastore, msgAndArgs...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mi.DMLMockFileSuffix()
			if !tt.wantErr(t, err, "DMLMockFileSuffix()") {
				return
			}
			assert.Equalf(t, tt.want, got, "DMLMockFileSuffix()")
		})
	}
}
