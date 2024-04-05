package monaco_test

import (
	"context"
	"testing"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository"
	. "code-platform/service/monaco"
	"code-platform/storage"

	"github.com/stretchr/testify/require"
)

func testHelper() (*storage.Storage, *MonacoService) {
	testStorage := testx.NewStorage()
	dao := &repository.Dao{Storage: testStorage}
	monacoService := NewMonacoService(dao, log.Sub("monaco"), NewMonacoClient())
	return testStorage, monacoService
}

func TestExecCode(t *testing.T) {
	testStorage, monacoService := testHelper()
	defer testStorage.Close()

	for _, c := range []struct {
		expectedError    error
		label            string
		code             string
		expectedResponse string
		maxTimeout       time.Duration
		language         int8
	}{
		{
			label:      "python single quote",
			language:   0,
			maxTimeout: time.Second * 30,
			code: `
print('hello world')
		`,
			expectedError:    nil,
			expectedResponse: "hello world\n",
		},
		{
			label:      "c++ alloc too much",
			language:   1,
			maxTimeout: time.Second * 30,
			code: `
#include <iostream>
using namespace std;

int main ()
{ 
	for (int i=0;i<100000;i++){
		int * a = new int[1000000];
	}
	cout << "Hello world." << endl;
	return 0;
}
`,
			expectedError: errorx.ErrOOMKilled,
		},
		{
			label:      "java exec wrong",
			language:   2,
			maxTimeout: time.Second * 30,
			code: `
public class Solution{
	public static void main(String ...args){
		System.out.println("hello java"
	}
}
`,
			expectedError: errorx.ErrWrongCode,
		},
		{
			label:      "python timeout",
			language:   0,
			maxTimeout: time.Second,
			code: `
import time
time.sleep(30)
print("hello python")`,
			expectedError: errorx.ErrContextCancel,
		},

		{
			label:      "python",
			language:   0,
			maxTimeout: time.Second * 30,
			code: `
print("hello python")`,
			expectedResponse: "hello python\n",
			expectedError:    nil,
		},
		{
			label:      "cpp",
			language:   1,
			maxTimeout: time.Second * 30,
			code: `
#include<iostream>
using namespace std;

int main(){
	cout<<"hello c++"<<endl;
	return 0;
}
`,
			expectedResponse: "hello c++\n",
			expectedError:    nil,
		},
		{
			label:      "java",
			language:   2,
			maxTimeout: time.Second * 30,
			code: `
public class Solution{
	public static void main(String ...args){
		System.out.println("hello java");
	}
}
`,
			expectedResponse: "hello java\n",
			expectedError:    nil,
		},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), c.maxTimeout)
		resp, err := monacoService.ExecCode(ctx, c.language, c.code)
		cancel()
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Equal(t, c.expectedResponse, resp, c.label)
		}
	}
}
