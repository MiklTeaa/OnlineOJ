package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"code-platform/api/grpc/ide/pb"
	"code-platform/pkg/filex"
	"code-platform/pkg/osx"
	"code-platform/service/ide/define"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *IDEServer) QuickViewCode(ctx context.Context, req *pb.QuickViewCodeRequest) (*pb.QuickViewCodeResponse, error) {
	codePath := filepath.Join(
		define.InitBasePath, "codespaces",
		fmt.Sprintf("workspace-%d", req.GetLabId()),
		strconv.FormatUint(req.GetUserId(), 10),
	)

	fileInfo, err := os.Stat(codePath)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return &pb.QuickViewCodeResponse{RootNode: &pb.QuickViewCodeResponse_FileNode{}}, nil
	default:
		return nil, status.Error(codes.Internal, err.Error())
	}

	root, err := buildTreeNodeForDir(codePath, fileInfo.Name())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.QuickViewCodeResponse{RootNode: root}, nil
}

func buildTreeNodeForDir(codePath string, name string) (*pb.QuickViewCodeResponse_FileNode, error) {
	childs, err := os.ReadDir(codePath)
	if err != nil {
		return nil, err
	}

	root := &pb.QuickViewCodeResponse_FileNode{
		Name:       name,
		IsDir:      true,
		ChildNodes: make([]*pb.QuickViewCodeResponse_FileNode, 0, len(childs)),
	}

	for _, child := range childs {
		childName := child.Name()
		if child.IsDir() {
			childNode, err := buildTreeNodeForDir(filepath.Join(codePath, childName), childName)
			if err != nil {
				return nil, err
			}
			root.ChildNodes = append(root.ChildNodes, childNode)
		} else {
			// 判断是否有后缀名
			childNameParts := strings.Split(childName, ".")
			if len(childNameParts) >= 2 {
				childFile, err := os.Open(filepath.Join(codePath, childName))
				if err != nil {
					return nil, err
				}

				// 避免直接读取二进制文件
				isBinary, err := filex.IsBinary(childFile)
				if err != nil {
					childFile.Close()
					return nil, err
				}

				var content string = "二进制文件"
				if !isBinary {
					data, err := io.ReadAll(childFile)
					if err != nil {
						childFile.Close()
						return nil, err
					}
					content = string(data)
				}
				childFile.Close()

				root.ChildNodes = append(root.ChildNodes, &pb.QuickViewCodeResponse_FileNode{
					Name:    childName,
					IsDir:   false,
					Content: content,
				})

			}
		}
	}
	return root, nil
}

func (i *IDEServer) GenerateTestFileForViewCode(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	pythonBuf := bytes.NewBufferString(`
	print("hello world")
	print("world hello")
`)
	cppBuf := bytes.NewBufferString(`
#include <iostream>
using namespace std;
int main(){
	cout<<"hello world"<<endl;
}
	`)
	javaBuf := bytes.NewBufferString(`public class Solution{
		public static void main(String ...args) {
			System.out.println("hello world");
		}
	}`)

	for i, buf := range []*bytes.Buffer{pythonBuf, cppBuf, javaBuf} {
		var fileName string
		switch i {
		case 0:
			fileName = "1.py"
		case 1:
			fileName = "1.cpp"
		case 2:
			fileName = "Solution.java"
		}

		filePath := filepath.Join(define.InitBasePath, "codespaces", "workspace-0", "0", fileName)
		err := osx.CreateFileIfNotExists(filePath)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		fileObj, err := os.OpenFile(filePath, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		_, err = fileObj.Write(buf.Bytes())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		fileObj.Close()
	}
	return &pb.Empty{}, nil
}

func (i *IDEServer) RemoveGenerateTestFileForViewCode(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	filePath := filepath.Join(define.InitBasePath, "codespaces", "workspace-0", "0")
	if err := os.RemoveAll(filePath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}
