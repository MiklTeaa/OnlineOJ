package pb;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.45.1)",
    comments = "Source: plagiarism_detection.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class plagiarismDetectionGrpc {

  private plagiarismDetectionGrpc() {}

  public static final String SERVICE_NAME = "plagiarism_detection.plagiarismDetection";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest,
      pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse> getDuplicateCheckMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DuplicateCheck",
      requestType = pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest.class,
      responseType = pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest,
      pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse> getDuplicateCheckMethod() {
    io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest, pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse> getDuplicateCheckMethod;
    if ((getDuplicateCheckMethod = plagiarismDetectionGrpc.getDuplicateCheckMethod) == null) {
      synchronized (plagiarismDetectionGrpc.class) {
        if ((getDuplicateCheckMethod = plagiarismDetectionGrpc.getDuplicateCheckMethod) == null) {
          plagiarismDetectionGrpc.getDuplicateCheckMethod = getDuplicateCheckMethod =
              io.grpc.MethodDescriptor.<pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest, pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DuplicateCheck"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse.getDefaultInstance()))
              .setSchemaDescriptor(new plagiarismDetectionMethodDescriptorSupplier("DuplicateCheck"))
              .build();
        }
      }
    }
    return getDuplicateCheckMethod;
  }

  private static volatile io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.ViewReportRequest,
      pb.plagiarismDetectionServerImplBase.ViewReportResponse> getViewReportMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ViewReport",
      requestType = pb.plagiarismDetectionServerImplBase.ViewReportRequest.class,
      responseType = pb.plagiarismDetectionServerImplBase.ViewReportResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.ViewReportRequest,
      pb.plagiarismDetectionServerImplBase.ViewReportResponse> getViewReportMethod() {
    io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.ViewReportRequest, pb.plagiarismDetectionServerImplBase.ViewReportResponse> getViewReportMethod;
    if ((getViewReportMethod = plagiarismDetectionGrpc.getViewReportMethod) == null) {
      synchronized (plagiarismDetectionGrpc.class) {
        if ((getViewReportMethod = plagiarismDetectionGrpc.getViewReportMethod) == null) {
          plagiarismDetectionGrpc.getViewReportMethod = getViewReportMethod =
              io.grpc.MethodDescriptor.<pb.plagiarismDetectionServerImplBase.ViewReportRequest, pb.plagiarismDetectionServerImplBase.ViewReportResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ViewReport"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.ViewReportRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.ViewReportResponse.getDefaultInstance()))
              .setSchemaDescriptor(new plagiarismDetectionMethodDescriptorSupplier("ViewReport"))
              .build();
        }
      }
    }
    return getViewReportMethod;
  }

  private static volatile io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest,
      pb.plagiarismDetectionServerImplBase.Empty> getGenerateTestFilesForDuplicateCheckMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GenerateTestFilesForDuplicateCheck",
      requestType = pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest.class,
      responseType = pb.plagiarismDetectionServerImplBase.Empty.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest,
      pb.plagiarismDetectionServerImplBase.Empty> getGenerateTestFilesForDuplicateCheckMethod() {
    io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest, pb.plagiarismDetectionServerImplBase.Empty> getGenerateTestFilesForDuplicateCheckMethod;
    if ((getGenerateTestFilesForDuplicateCheckMethod = plagiarismDetectionGrpc.getGenerateTestFilesForDuplicateCheckMethod) == null) {
      synchronized (plagiarismDetectionGrpc.class) {
        if ((getGenerateTestFilesForDuplicateCheckMethod = plagiarismDetectionGrpc.getGenerateTestFilesForDuplicateCheckMethod) == null) {
          plagiarismDetectionGrpc.getGenerateTestFilesForDuplicateCheckMethod = getGenerateTestFilesForDuplicateCheckMethod =
              io.grpc.MethodDescriptor.<pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest, pb.plagiarismDetectionServerImplBase.Empty>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GenerateTestFilesForDuplicateCheck"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.Empty.getDefaultInstance()))
              .setSchemaDescriptor(new plagiarismDetectionMethodDescriptorSupplier("GenerateTestFilesForDuplicateCheck"))
              .build();
        }
      }
    }
    return getGenerateTestFilesForDuplicateCheckMethod;
  }

  private static volatile io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.Empty,
      pb.plagiarismDetectionServerImplBase.Empty> getRemoveTestFilesForDuplicateCheckMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RemoveTestFilesForDuplicateCheck",
      requestType = pb.plagiarismDetectionServerImplBase.Empty.class,
      responseType = pb.plagiarismDetectionServerImplBase.Empty.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.Empty,
      pb.plagiarismDetectionServerImplBase.Empty> getRemoveTestFilesForDuplicateCheckMethod() {
    io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.Empty, pb.plagiarismDetectionServerImplBase.Empty> getRemoveTestFilesForDuplicateCheckMethod;
    if ((getRemoveTestFilesForDuplicateCheckMethod = plagiarismDetectionGrpc.getRemoveTestFilesForDuplicateCheckMethod) == null) {
      synchronized (plagiarismDetectionGrpc.class) {
        if ((getRemoveTestFilesForDuplicateCheckMethod = plagiarismDetectionGrpc.getRemoveTestFilesForDuplicateCheckMethod) == null) {
          plagiarismDetectionGrpc.getRemoveTestFilesForDuplicateCheckMethod = getRemoveTestFilesForDuplicateCheckMethod =
              io.grpc.MethodDescriptor.<pb.plagiarismDetectionServerImplBase.Empty, pb.plagiarismDetectionServerImplBase.Empty>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RemoveTestFilesForDuplicateCheck"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.Empty.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.Empty.getDefaultInstance()))
              .setSchemaDescriptor(new plagiarismDetectionMethodDescriptorSupplier("RemoveTestFilesForDuplicateCheck"))
              .build();
        }
      }
    }
    return getRemoveTestFilesForDuplicateCheckMethod;
  }

  private static volatile io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest,
      pb.plagiarismDetectionServerImplBase.Empty> getGenerateTestHTMLFileForViewReportMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GenerateTestHTMLFileForViewReport",
      requestType = pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest.class,
      responseType = pb.plagiarismDetectionServerImplBase.Empty.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest,
      pb.plagiarismDetectionServerImplBase.Empty> getGenerateTestHTMLFileForViewReportMethod() {
    io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest, pb.plagiarismDetectionServerImplBase.Empty> getGenerateTestHTMLFileForViewReportMethod;
    if ((getGenerateTestHTMLFileForViewReportMethod = plagiarismDetectionGrpc.getGenerateTestHTMLFileForViewReportMethod) == null) {
      synchronized (plagiarismDetectionGrpc.class) {
        if ((getGenerateTestHTMLFileForViewReportMethod = plagiarismDetectionGrpc.getGenerateTestHTMLFileForViewReportMethod) == null) {
          plagiarismDetectionGrpc.getGenerateTestHTMLFileForViewReportMethod = getGenerateTestHTMLFileForViewReportMethod =
              io.grpc.MethodDescriptor.<pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest, pb.plagiarismDetectionServerImplBase.Empty>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GenerateTestHTMLFileForViewReport"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.Empty.getDefaultInstance()))
              .setSchemaDescriptor(new plagiarismDetectionMethodDescriptorSupplier("GenerateTestHTMLFileForViewReport"))
              .build();
        }
      }
    }
    return getGenerateTestHTMLFileForViewReportMethod;
  }

  private static volatile io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.Empty,
      pb.plagiarismDetectionServerImplBase.Empty> getRemoveTestHTMLFileForViewReportMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RemoveTestHTMLFileForViewReport",
      requestType = pb.plagiarismDetectionServerImplBase.Empty.class,
      responseType = pb.plagiarismDetectionServerImplBase.Empty.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.Empty,
      pb.plagiarismDetectionServerImplBase.Empty> getRemoveTestHTMLFileForViewReportMethod() {
    io.grpc.MethodDescriptor<pb.plagiarismDetectionServerImplBase.Empty, pb.plagiarismDetectionServerImplBase.Empty> getRemoveTestHTMLFileForViewReportMethod;
    if ((getRemoveTestHTMLFileForViewReportMethod = plagiarismDetectionGrpc.getRemoveTestHTMLFileForViewReportMethod) == null) {
      synchronized (plagiarismDetectionGrpc.class) {
        if ((getRemoveTestHTMLFileForViewReportMethod = plagiarismDetectionGrpc.getRemoveTestHTMLFileForViewReportMethod) == null) {
          plagiarismDetectionGrpc.getRemoveTestHTMLFileForViewReportMethod = getRemoveTestHTMLFileForViewReportMethod =
              io.grpc.MethodDescriptor.<pb.plagiarismDetectionServerImplBase.Empty, pb.plagiarismDetectionServerImplBase.Empty>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RemoveTestHTMLFileForViewReport"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.Empty.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  pb.plagiarismDetectionServerImplBase.Empty.getDefaultInstance()))
              .setSchemaDescriptor(new plagiarismDetectionMethodDescriptorSupplier("RemoveTestHTMLFileForViewReport"))
              .build();
        }
      }
    }
    return getRemoveTestHTMLFileForViewReportMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static plagiarismDetectionStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<plagiarismDetectionStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<plagiarismDetectionStub>() {
        @java.lang.Override
        public plagiarismDetectionStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new plagiarismDetectionStub(channel, callOptions);
        }
      };
    return plagiarismDetectionStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static plagiarismDetectionBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<plagiarismDetectionBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<plagiarismDetectionBlockingStub>() {
        @java.lang.Override
        public plagiarismDetectionBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new plagiarismDetectionBlockingStub(channel, callOptions);
        }
      };
    return plagiarismDetectionBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static plagiarismDetectionFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<plagiarismDetectionFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<plagiarismDetectionFutureStub>() {
        @java.lang.Override
        public plagiarismDetectionFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new plagiarismDetectionFutureStub(channel, callOptions);
        }
      };
    return plagiarismDetectionFutureStub.newStub(factory, channel);
  }

  /**
   */
  public static abstract class plagiarismDetectionImplBase implements io.grpc.BindableService {

    /**
     */
    public void duplicateCheck(pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDuplicateCheckMethod(), responseObserver);
    }

    /**
     */
    public void viewReport(pb.plagiarismDetectionServerImplBase.ViewReportRequest request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.ViewReportResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getViewReportMethod(), responseObserver);
    }

    /**
     * <pre>
     * GenerateTestFiles 生成代码文件以作测试用
     * </pre>
     */
    public void generateTestFilesForDuplicateCheck(pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGenerateTestFilesForDuplicateCheckMethod(), responseObserver);
    }

    /**
     */
    public void removeTestFilesForDuplicateCheck(pb.plagiarismDetectionServerImplBase.Empty request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRemoveTestFilesForDuplicateCheckMethod(), responseObserver);
    }

    /**
     * <pre>
     * GenerateTestFiles 生成HTML文件以作测试用
     * </pre>
     */
    public void generateTestHTMLFileForViewReport(pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGenerateTestHTMLFileForViewReportMethod(), responseObserver);
    }

    /**
     */
    public void removeTestHTMLFileForViewReport(pb.plagiarismDetectionServerImplBase.Empty request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRemoveTestHTMLFileForViewReportMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getDuplicateCheckMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest,
                pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse>(
                  this, METHODID_DUPLICATE_CHECK)))
          .addMethod(
            getViewReportMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                pb.plagiarismDetectionServerImplBase.ViewReportRequest,
                pb.plagiarismDetectionServerImplBase.ViewReportResponse>(
                  this, METHODID_VIEW_REPORT)))
          .addMethod(
            getGenerateTestFilesForDuplicateCheckMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest,
                pb.plagiarismDetectionServerImplBase.Empty>(
                  this, METHODID_GENERATE_TEST_FILES_FOR_DUPLICATE_CHECK)))
          .addMethod(
            getRemoveTestFilesForDuplicateCheckMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                pb.plagiarismDetectionServerImplBase.Empty,
                pb.plagiarismDetectionServerImplBase.Empty>(
                  this, METHODID_REMOVE_TEST_FILES_FOR_DUPLICATE_CHECK)))
          .addMethod(
            getGenerateTestHTMLFileForViewReportMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest,
                pb.plagiarismDetectionServerImplBase.Empty>(
                  this, METHODID_GENERATE_TEST_HTMLFILE_FOR_VIEW_REPORT)))
          .addMethod(
            getRemoveTestHTMLFileForViewReportMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                pb.plagiarismDetectionServerImplBase.Empty,
                pb.plagiarismDetectionServerImplBase.Empty>(
                  this, METHODID_REMOVE_TEST_HTMLFILE_FOR_VIEW_REPORT)))
          .build();
    }
  }

  /**
   */
  public static final class plagiarismDetectionStub extends io.grpc.stub.AbstractAsyncStub<plagiarismDetectionStub> {
    private plagiarismDetectionStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected plagiarismDetectionStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new plagiarismDetectionStub(channel, callOptions);
    }

    /**
     */
    public void duplicateCheck(pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDuplicateCheckMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void viewReport(pb.plagiarismDetectionServerImplBase.ViewReportRequest request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.ViewReportResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getViewReportMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GenerateTestFiles 生成代码文件以作测试用
     * </pre>
     */
    public void generateTestFilesForDuplicateCheck(pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGenerateTestFilesForDuplicateCheckMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void removeTestFilesForDuplicateCheck(pb.plagiarismDetectionServerImplBase.Empty request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRemoveTestFilesForDuplicateCheckMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * GenerateTestFiles 生成HTML文件以作测试用
     * </pre>
     */
    public void generateTestHTMLFileForViewReport(pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGenerateTestHTMLFileForViewReportMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void removeTestHTMLFileForViewReport(pb.plagiarismDetectionServerImplBase.Empty request,
        io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRemoveTestHTMLFileForViewReportMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   */
  public static final class plagiarismDetectionBlockingStub extends io.grpc.stub.AbstractBlockingStub<plagiarismDetectionBlockingStub> {
    private plagiarismDetectionBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected plagiarismDetectionBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new plagiarismDetectionBlockingStub(channel, callOptions);
    }

    /**
     */
    public pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse duplicateCheck(pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDuplicateCheckMethod(), getCallOptions(), request);
    }

    /**
     */
    public pb.plagiarismDetectionServerImplBase.ViewReportResponse viewReport(pb.plagiarismDetectionServerImplBase.ViewReportRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getViewReportMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GenerateTestFiles 生成代码文件以作测试用
     * </pre>
     */
    public pb.plagiarismDetectionServerImplBase.Empty generateTestFilesForDuplicateCheck(pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGenerateTestFilesForDuplicateCheckMethod(), getCallOptions(), request);
    }

    /**
     */
    public pb.plagiarismDetectionServerImplBase.Empty removeTestFilesForDuplicateCheck(pb.plagiarismDetectionServerImplBase.Empty request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRemoveTestFilesForDuplicateCheckMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * GenerateTestFiles 生成HTML文件以作测试用
     * </pre>
     */
    public pb.plagiarismDetectionServerImplBase.Empty generateTestHTMLFileForViewReport(pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGenerateTestHTMLFileForViewReportMethod(), getCallOptions(), request);
    }

    /**
     */
    public pb.plagiarismDetectionServerImplBase.Empty removeTestHTMLFileForViewReport(pb.plagiarismDetectionServerImplBase.Empty request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRemoveTestHTMLFileForViewReportMethod(), getCallOptions(), request);
    }
  }

  /**
   */
  public static final class plagiarismDetectionFutureStub extends io.grpc.stub.AbstractFutureStub<plagiarismDetectionFutureStub> {
    private plagiarismDetectionFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected plagiarismDetectionFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new plagiarismDetectionFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse> duplicateCheck(
        pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDuplicateCheckMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<pb.plagiarismDetectionServerImplBase.ViewReportResponse> viewReport(
        pb.plagiarismDetectionServerImplBase.ViewReportRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getViewReportMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GenerateTestFiles 生成代码文件以作测试用
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<pb.plagiarismDetectionServerImplBase.Empty> generateTestFilesForDuplicateCheck(
        pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGenerateTestFilesForDuplicateCheckMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<pb.plagiarismDetectionServerImplBase.Empty> removeTestFilesForDuplicateCheck(
        pb.plagiarismDetectionServerImplBase.Empty request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRemoveTestFilesForDuplicateCheckMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * GenerateTestFiles 生成HTML文件以作测试用
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<pb.plagiarismDetectionServerImplBase.Empty> generateTestHTMLFileForViewReport(
        pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGenerateTestHTMLFileForViewReportMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<pb.plagiarismDetectionServerImplBase.Empty> removeTestHTMLFileForViewReport(
        pb.plagiarismDetectionServerImplBase.Empty request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRemoveTestHTMLFileForViewReportMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_DUPLICATE_CHECK = 0;
  private static final int METHODID_VIEW_REPORT = 1;
  private static final int METHODID_GENERATE_TEST_FILES_FOR_DUPLICATE_CHECK = 2;
  private static final int METHODID_REMOVE_TEST_FILES_FOR_DUPLICATE_CHECK = 3;
  private static final int METHODID_GENERATE_TEST_HTMLFILE_FOR_VIEW_REPORT = 4;
  private static final int METHODID_REMOVE_TEST_HTMLFILE_FOR_VIEW_REPORT = 5;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final plagiarismDetectionImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(plagiarismDetectionImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_DUPLICATE_CHECK:
          serviceImpl.duplicateCheck((pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest) request,
              (io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse>) responseObserver);
          break;
        case METHODID_VIEW_REPORT:
          serviceImpl.viewReport((pb.plagiarismDetectionServerImplBase.ViewReportRequest) request,
              (io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.ViewReportResponse>) responseObserver);
          break;
        case METHODID_GENERATE_TEST_FILES_FOR_DUPLICATE_CHECK:
          serviceImpl.generateTestFilesForDuplicateCheck((pb.plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest) request,
              (io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty>) responseObserver);
          break;
        case METHODID_REMOVE_TEST_FILES_FOR_DUPLICATE_CHECK:
          serviceImpl.removeTestFilesForDuplicateCheck((pb.plagiarismDetectionServerImplBase.Empty) request,
              (io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty>) responseObserver);
          break;
        case METHODID_GENERATE_TEST_HTMLFILE_FOR_VIEW_REPORT:
          serviceImpl.generateTestHTMLFileForViewReport((pb.plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest) request,
              (io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty>) responseObserver);
          break;
        case METHODID_REMOVE_TEST_HTMLFILE_FOR_VIEW_REPORT:
          serviceImpl.removeTestHTMLFileForViewReport((pb.plagiarismDetectionServerImplBase.Empty) request,
              (io.grpc.stub.StreamObserver<pb.plagiarismDetectionServerImplBase.Empty>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  private static abstract class plagiarismDetectionBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    plagiarismDetectionBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return pb.plagiarismDetectionServerImplBase.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("plagiarismDetection");
    }
  }

  private static final class plagiarismDetectionFileDescriptorSupplier
      extends plagiarismDetectionBaseDescriptorSupplier {
    plagiarismDetectionFileDescriptorSupplier() {}
  }

  private static final class plagiarismDetectionMethodDescriptorSupplier
      extends plagiarismDetectionBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    plagiarismDetectionMethodDescriptorSupplier(String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (plagiarismDetectionGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new plagiarismDetectionFileDescriptorSupplier())
              .addMethod(getDuplicateCheckMethod())
              .addMethod(getViewReportMethod())
              .addMethod(getGenerateTestFilesForDuplicateCheckMethod())
              .addMethod(getRemoveTestFilesForDuplicateCheckMethod())
              .addMethod(getGenerateTestHTMLFileForViewReportMethod())
              .addMethod(getRemoveTestHTMLFileForViewReportMethod())
              .build();
        }
      }
    }
    return result;
  }
}
