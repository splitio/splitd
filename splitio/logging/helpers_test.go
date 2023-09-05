package logging

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetWriterStd(t *testing.T) {
	// stdout/stderr
	writer, err := GetWriter(ref("stdout"), nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, os.Stdout, writer)
	writer, err = GetWriter(ref("/dev/stdout"), nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, os.Stdout, writer)
	writer, err = GetWriter(ref("stderr"), nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, os.Stderr, writer)
	writer, err = GetWriter(ref("/dev/stderr"), nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, os.Stderr, writer)
}

func TestGetWriterFileNoRotation(t *testing.T) {
	// regular file, no rotation
	fn := strings.Join([]string{os.TempDir(), "someLogFile2"}, string(os.PathSeparator))
	defer func() {
		if err := os.Remove(fn); err != nil {
			t.Error("error deleting file: ", err)
		}
	}()
	writer, err := GetWriter(&fn, nil, nil)
	assert.Nil(t, err)

	writer.Write([]byte("hola que tal"))
	writer.(io.Closer).Close()

	assertFileContents(t, "hola que tal", fn)
}

func TestGetWriterFileRotation(t *testing.T) {

    // remove any old file
    fl, err := os.ReadDir(os.TempDir())
    assert.Nil(t, err)
    for _, fe := range fl {
        if strings.HasPrefix(fe.Name(), "someLogFile") {
            os.Remove(fe.Name())
        }
    }


	fn := strings.Join([]string{os.TempDir(), "someLogFile"}, string(os.PathSeparator))
	writer, err := GetWriter(&fn, ref(2), ref(5))
	assert.Nil(t, err)

	writer.Write([]byte("12345"))
	writer.Write([]byte("67890"))
	writer.Write([]byte("qwert"))
	writer.Write([]byte("asdfg"))
	writer.Write([]byte("zxcvb"))
	writer.Write([]byte("hjaiu"))


    time.Sleep(1*time.Second) // file rotate writer is async.. give it a second before looking at the fs
    fl, err = os.ReadDir(os.TempDir())
    assert.Nil(t, err)
    names := make([]string, 0, 3)
    for _, fe := range fl {
        if strings.HasPrefix(fe.Name(), "someLogFile") {
            names = append(names, fe.Name())
        }
    }

    assert.Contains(t, names, "someLogFile")
    assert.Contains(t, names, "someLogFile.1")
    assert.Contains(t, names, "someLogFile.2")
    assert.NotContains(t, names, "someLogFile.3")
}

func assertFileContents(t *testing.T, expected string, fn string) {
	t.Helper()
	contents, err := ioutil.ReadFile(fn)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(contents))
}
func ref[T any](t T) *T {
	return &t
}
