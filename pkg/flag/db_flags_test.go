package flag_test

import (
	"github.com/google/go-containerregistry/pkg/name"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/aquasecurity/trivy/pkg/flag"
	"github.com/aquasecurity/trivy/pkg/log"
)

func TestDBFlagGroup_ToOptions(t *testing.T) {
	type fields struct {
		SkipDBUpdate     bool
		DownloadDBOnly   bool
		Light            bool
		DBRepository     string
		JavaDBRepository string
	}
	tests := []struct {
		name      string
		fields    fields
		want      flag.DBOptions
		wantLogs  []string
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "happy",
			fields: fields{
				SkipDBUpdate:     true,
				DownloadDBOnly:   false,
				DBRepository:     "ghcr.io/aquasecurity/trivy-db",
				JavaDBRepository: "ghcr.io/aquasecurity/trivy-java-db",
			},
			want: flag.DBOptions{
				SkipDBUpdate:     true,
				DownloadDBOnly:   false,
				DBRepository:     name.Tag{}, // All fields are unexported
				JavaDBRepository: name.Tag{}, // All fields are unexported
			},
			assertion: require.NoError,
		},
		{
			name: "light",
			fields: fields{
				Light:            true,
				DBRepository:     "ghcr.io/aquasecurity/trivy-db",
				JavaDBRepository: "ghcr.io/aquasecurity/trivy-java-db",
			},
			want: flag.DBOptions{
				Light:            true,
				DBRepository:     name.Tag{}, // All fields are unexported
				JavaDBRepository: name.Tag{}, // All fields are unexported
			},
			wantLogs: []string{
				"'--light' option is deprecated and will be removed. See also: https://github.com/aquasecurity/trivy/discussions/1649",
			},
			assertion: require.NoError,
		},
		{
			name: "sad",
			fields: fields{
				SkipDBUpdate:   true,
				DownloadDBOnly: true,
			},
			assertion: func(t require.TestingT, err error, msgs ...interface{}) {
				require.ErrorContains(t, err, "--skip-db-update and --download-db-only options can not be specified both")
			},
		},
		{
			name: "invalid repo",
			fields: fields{
				SkipDBUpdate:   true,
				DownloadDBOnly: false,
				DBRepository:   "foo:bar:baz",
			},
			assertion: func(t require.TestingT, err error, msgs ...interface{}) {
				require.ErrorContains(t, err, "invalid db repository")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := zap.WarnLevel
			core, obs := observer.New(level)
			log.Logger = zap.New(core).Sugar()

			viper.Set(flag.SkipDBUpdateFlag.ConfigName, tt.fields.SkipDBUpdate)
			viper.Set(flag.DownloadDBOnlyFlag.ConfigName, tt.fields.DownloadDBOnly)
			viper.Set(flag.LightFlag.ConfigName, tt.fields.Light)
			viper.Set(flag.DBRepositoryFlag.ConfigName, tt.fields.DBRepository)
			viper.Set(flag.JavaDBRepositoryFlag.ConfigName, tt.fields.JavaDBRepository)

			// Assert options
			f := &flag.DBFlagGroup{
				DownloadDBOnly:   flag.DownloadDBOnlyFlag.Clone(),
				SkipDBUpdate:     flag.SkipDBUpdateFlag.Clone(),
				Light:            flag.LightFlag.Clone(),
				DBRepository:     flag.DBRepositoryFlag.Clone(),
				JavaDBRepository: flag.JavaDBRepositoryFlag.Clone(),
			}
			got, err := f.ToOptions()
			tt.assertion(t, err)
			assert.EqualExportedValues(t, tt.want, got)

			// Assert log messages
			var gotMessages []string
			for _, entry := range obs.AllUntimed() {
				gotMessages = append(gotMessages, entry.Message)
			}
			assert.Equal(t, tt.wantLogs, gotMessages, tt.name)
		})
	}
}
