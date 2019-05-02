package restore_test

/*
 * This file contains integration tests for gprestore as a whole, rather than
 * tests relating to functions in any particular file.
 */

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gpbackup/restore"
	"github.com/greenplum-db/gpbackup/testutils"
	"github.com/greenplum-db/gpbackup/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/spf13/pflag"
)

var (
	connectionPool *dbconn.DBConn
	mock           sqlmock.Sqlmock
	stdout         *gbytes.Buffer
	stderr         *gbytes.Buffer
	logfile        *gbytes.Buffer
	buffer         *gbytes.Buffer
)

func TestRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "restore tests")
}

var cmdFlags *pflag.FlagSet

var _ = BeforeEach(func() {
	connectionPool, mock, stdout, stderr, logfile = testutils.SetupTestEnvironment()
	restore.SetConnection(connectionPool)
	buffer = gbytes.NewBuffer()

	cmdFlags = pflag.NewFlagSet("gprestore", pflag.ExitOnError)
	restore.SetCmdFlags(cmdFlags)

	cmdFlags.Bool(utils.ON_ERROR_CONTINUE, false, "")
	cmdFlags.Bool(utils.DATA_ONLY, false, "")
	cmdFlags.String(utils.PLUGIN_CONFIG, "", "")
	cmdFlags.StringSlice(utils.INCLUDE_RELATION, []string{}, "")
	cmdFlags.StringSlice(utils.EXCLUDE_RELATION, []string{}, "")
	cmdFlags.StringSlice(utils.INCLUDE_SCHEMA, []string{}, "")
	cmdFlags.StringSlice(utils.EXCLUDE_SCHEMA, []string{}, "")
})
