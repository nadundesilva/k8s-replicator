/*
 * Copyright (c) 2022, Nadun De Silva. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *   http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

type mockReader struct {
	output       []byte
	err          error
	currentIndex int
}

func (r *mockReader) Read(buffer []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}

	expectedReadSize := len(buffer)
	readSize := 0
	for readSize < expectedReadSize && r.currentIndex < len(r.output) {
		buffer[readSize] = r.output[r.currentIndex]
		readSize++
		r.currentIndex++
	}

	if readSize == 0 {
		return 0, io.EOF
	}
	return readSize, r.err
}

func TestReplaceEnv(t *testing.T) {
	os.Setenv("K8S_REPLICATOR_TEST_ENV_KEY", "K8S_REPLICATOR_TEST_ENV_VALUE")

	type testDatum struct {
		name           string
		reader         *mockReader
		expectedString string
		expectedErr    error
	}
	testData := []testDatum{
		{
			name: "Env var expansion with env var",
			reader: &mockReader{
				output: []byte("This is a sample with env variable \"${K8S_REPLICATOR_TEST_ENV_KEY}\""),
				err:    nil,
			},
			expectedString: "This is a sample with env variable \"K8S_REPLICATOR_TEST_ENV_VALUE\"",
			expectedErr:    nil,
		},
		{
			name: "Env var expansion without env var",
			reader: &mockReader{
				output: []byte("This is a sample without env variable"),
				err:    nil,
			},
			expectedString: "This is a sample without env variable",
			expectedErr:    nil,
		},
		{
			name: "Env var expansion with empty string",
			reader: &mockReader{
				output: []byte(""),
				err:    nil,
			},
			expectedString: "",
			expectedErr:    nil,
		},
		{
			name: "Env var expansion with io.Reader returning error",
			reader: &mockReader{
				output: nil,
				err:    fmt.Errorf("Test reader error"),
			},
			expectedString: "",
			expectedErr:    fmt.Errorf("Test reader error"),
		},
		{
			name: "Env var expansion with missing env var",
			reader: &mockReader{
				output: []byte("This is a sample with env variable \"${K8S_REPLICATOR_MISSING_TEST_ENV_KEY}\""),
				err:    nil,
			},
			expectedString: "This is a sample with env variable \"K8S_REPLICATOR_TEST_ENV_VALUE\"",
			expectedErr:    fmt.Errorf("missing env(s): K8S_REPLICATOR_MISSING_TEST_ENV_KEY"),
		},
		{
			name: "Env var expansion with multiple missing env vars",
			reader: &mockReader{
				output: []byte("This is a sample with env variables \"${K8S_REPLICATOR_MISSING_TEST_ENV_KEY}\" and \"${K8S_REPLICATOR_MISSING_TEST_ENV_KEY_2}\""),
				err:    nil,
			},
			expectedString: "This is a sample with env variable \"K8S_REPLICATOR_TEST_ENV_VALUE\"",
			expectedErr:    fmt.Errorf("missing env(s): K8S_REPLICATOR_MISSING_TEST_ENV_KEY, K8S_REPLICATOR_MISSING_TEST_ENV_KEY_2"),
		},
	}
	for _, testDatum := range testData {
		t.Run(testDatum.name, func(t *testing.T) {
			replacedReader, err := replaceEnv(testDatum.reader)
			if testDatum.expectedErr != nil && err == nil {
				t.Errorf("expected error not returned: %v", testDatum.expectedErr)
			}
			if testDatum.expectedErr == nil && err != nil {
				t.Errorf("unexpected error returned: %v", err)
			}
			if testDatum.expectedErr != nil && err != nil {
				if testDatum.expectedErr.Error() != err.Error() {
					t.Errorf("returned invalid error: want %v; got %v", testDatum.expectedErr, err)
				}
				if replacedReader != nil {
					t.Errorf("expected returned reader to be nil, but found %+v", replacedReader)
				}
			} else {
				replacedBytes, err := ioutil.ReadAll(replacedReader)
				if err != nil {
					t.Errorf("returned io.Reader returned error: %v", err)
				}
				if string(replacedBytes) != testDatum.expectedString {
					t.Errorf("returned io.Returned unexpected value: want `%s`; got `%s`",
						testDatum.expectedString, replacedBytes)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	type testDatum struct {
		name           string
		reader         *mockReader
		configType     string
		expectedConfig *Conf
		expectedErr    error
	}
	testData := []testDatum{
		{
			name: "Config read (yaml)",
			reader: &mockReader{
				output: []byte("apiVersion: replicator.nadundesilva.github.io/v1\n" +
					"kind: Config\n" +
					"logging:\n" +
					"  level: \"debug\"\n" +
					"resources:\n" +
					"  - apiVersion: v1\n" +
					"    kind: Secret\n" +
					"  - apiVersion: networking.k8s.io/v1\n" +
					"    kind: NetworkPolicy\n",
				),
				err: nil,
			},
			configType: "yaml",
			expectedConfig: &Conf{
				Logging: LoggingConf{
					Level: "debug",
				},
				Resources: []ResourceType{
					{
						ApiVersion: "v1",
						Kind:       "Secret",
					},
					{
						ApiVersion: "networking.k8s.io/v1",
						Kind:       "NetworkPolicy",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Config read with io.Reader failing",
			reader: &mockReader{
				output: nil,
				err:    fmt.Errorf("Test reader error for replace env"),
			},
			configType:     "yaml",
			expectedConfig: nil,
			expectedErr:    fmt.Errorf("Test reader error for replace env"),
		},
		{
			name: "Config read with io.Reader output not matching config type",
			reader: &mockReader{
				output: []byte("{}"),
				err:    nil,
			},
			configType:     "ini",
			expectedConfig: nil,
			expectedErr:    fmt.Errorf("While parsing config: key-value delimiter not found: {}"),
		},
		{
			name: "Config read with invalid api version",
			reader: &mockReader{
				output: []byte("apiVersion: invalid\n" +
					"kind: Config\n" +
					"logging:\n" +
					"  level: \"debug\"\n" +
					"resources:\n" +
					"  - apiVersion: v1\n" +
					"    kind: Secret\n",
				),
				err: nil,
			},
			configType:     "yaml",
			expectedConfig: nil,
			expectedErr:    fmt.Errorf("invalid config file api version invalid, expected replicator.nadundesilva.github.io/v1"),
		},
		{
			name: "Config read with invalid kind",
			reader: &mockReader{
				output: []byte("apiVersion: replicator.nadundesilva.github.io/v1\n" +
					"kind: Invalid\n" +
					"logging:\n" +
					"  level: \"debug\"\n" +
					"resources:\n" +
					"  - apiVersion: v1\n" +
					"    kind: Secret\n",
				),
				err: nil,
			},
			configType:     "yaml",
			expectedConfig: nil,
			expectedErr:    fmt.Errorf("invalid config file kind Invalid, expected Config"),
		},
		{
			name: "Config read (yaml) with invalid io.Reader output",
			reader: &mockReader{
				output: []byte("apiVersion: replicator.nadundesilva.github.io/v1\n" +
					"kind: Config\n" +
					"logging:\n" +
					"  level: \"debug\"\n" +
					"resources: foo",
				),
				err: nil,
			},
			configType:     "yaml",
			expectedConfig: nil,
			expectedErr:    fmt.Errorf("1 error(s) decoding:\n\n* 'resources[0]' expected a map, got 'string'"),
		},
	}
	for _, testDatum := range testData {
		t.Run(testDatum.name, func(t *testing.T) {
			conf, err := New(testDatum.reader, testDatum.configType)
			if testDatum.expectedErr != nil && err == nil {
				t.Errorf("expected error not returned: %v", testDatum.expectedErr)
			}
			if testDatum.expectedErr == nil && err != nil {
				t.Errorf("unexpected error returned: %v", err)
			}
			if testDatum.expectedErr != nil && err != nil {
				if testDatum.expectedErr.Error() != err.Error() {
					t.Errorf("returned invalid error: want `%v`; got `%v`", testDatum.expectedErr, err)
				}
				if conf != nil {
					t.Errorf("expected config to be nil, but found %+v", conf)
				}
			} else {
				if !reflect.DeepEqual(testDatum.expectedConfig, conf) {
					t.Errorf("unexpected parsed config: want %+v; got %+v", testDatum.expectedConfig, conf)
				}
			}
		})
	}
}

func TestNewFromFile(t *testing.T) {
	type testDatum struct {
		name           string
		confFile       string
		expectedConfig *Conf
		expectedErr    error
	}
	conf := &Conf{
		Logging: LoggingConf{
			Level: "debug",
		},
		Resources: []ResourceType{
			{
				ApiVersion: "v1",
				Kind:       "Secret",
			},
			{
				ApiVersion: "v1",
				Kind:       "ConfigMap",
			},
			{
				ApiVersion: "networking.k8s.io/v1",
				Kind:       "NetworkPolicy",
			},
		},
	}
	testData := []testDatum{
		{
			name:           "Config read from yaml file",
			confFile:       "testdata/conf.yaml",
			expectedConfig: conf,
			expectedErr:    nil,
		},
		{
			name:           "Config read from yml file",
			confFile:       "testdata/conf.yml",
			expectedConfig: conf,
			expectedErr:    nil,
		},
		{
			name:           "Config read from json file",
			confFile:       "testdata/conf.json",
			expectedConfig: conf,
			expectedErr:    nil,
		},
		{
			name:           "Config read from toml file",
			confFile:       "testdata/conf.toml",
			expectedConfig: conf,
			expectedErr:    nil,
		},
		{
			name:           "Config read from file without extension",
			confFile:       "testdata/conf",
			expectedConfig: conf,
			expectedErr:    nil,
		},
		{
			name:           "Config read from non existent file",
			confFile:       "testdata/invalid.yaml",
			expectedConfig: nil,
			expectedErr:    fmt.Errorf("open testdata/invalid.yaml: no such file or directory"),
		},
	}
	for _, testDatum := range testData {
		t.Run(testDatum.name, func(t *testing.T) {
			conf, err := NewFromFile(testDatum.confFile)
			if testDatum.expectedErr != nil && err == nil {
				t.Errorf("expected error not returned: %v", testDatum.expectedErr)
			}
			if testDatum.expectedErr == nil && err != nil {
				t.Errorf("unexpected error returned: %v", err)
			}
			if testDatum.expectedErr != nil && err != nil {
				if testDatum.expectedErr.Error() != err.Error() {
					t.Errorf("returned invalid error: want `%v`; got `%v`", testDatum.expectedErr, err)
				}
				if conf != nil {
					t.Errorf("expected config to be nil, but found %+v", conf)
				}
			} else {
				if !reflect.DeepEqual(testDatum.expectedConfig, conf) {
					t.Errorf("unexpected parsed config: want %+v; got %+v", testDatum.expectedConfig, conf)
				}
			}
		})
	}
}
