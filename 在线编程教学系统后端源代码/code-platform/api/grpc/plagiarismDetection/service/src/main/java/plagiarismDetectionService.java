import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileReader;
import java.io.FileWriter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Arrays;
import java.util.List;

import de.jplag.JPlag;
import de.jplag.JPlagComparison;
import de.jplag.JPlagResult;
import de.jplag.exceptions.ReportGenerationException;
import de.jplag.exceptions.RootDirectoryException;
import de.jplag.exceptions.SubmissionException;
import de.jplag.options.JPlagOptions;
import de.jplag.options.LanguageOption;
import de.jplag.reporting.Report;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import pb.plagiarismDetectionGrpc.plagiarismDetectionImplBase;
import pb.plagiarismDetectionServerImplBase;
import pb.plagiarismDetectionServerImplBase.DuplicateCheckRequest;
import pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse;
import pb.plagiarismDetectionServerImplBase.DuplicateCheckResponse.DuplicateCheckResponseValue;
import pb.plagiarismDetectionServerImplBase.Empty;
import pb.plagiarismDetectionServerImplBase.ViewReportRequest;
import pb.plagiarismDetectionServerImplBase.ViewReportResponse;
import pb.plagiarismDetectionServerImplBase.language;

public class plagiarismDetectionService extends plagiarismDetectionImplBase {
	private static final String outputPathPrefix = "/results";
	private static final String htmlFileLayout = "match%d.html";
	private static final String baseCodePath = "/code_platform/workspace/codespaces";

	private LanguageOption getLanguageType(language lan) {
		if (lan == language.python3) {
			return LanguageOption.PYTHON_3;
		} else if (lan == language.cpp) {
			return LanguageOption.C_CPP;
		} else {
			return LanguageOption.JAVA;
		}
	}

	@Override
	public void duplicateCheck(DuplicateCheckRequest request,
							   StreamObserver<DuplicateCheckResponse> responseObserver) {

		Path codePath = Paths.get(baseCodePath, String.format("workspace-%d", request.getLabID()));
		if (!Files.exists(codePath)) {
			responseObserver.onError(Status.NOT_FOUND.withDescription("lab dir is not found").asException());
			return;
		}

		JPlagOptions options = new JPlagOptions(codePath.toString(), getLanguageType(request.getLan()));
		// TODO 处理空文件
		options.setMinimumTokenMatch(1);
		try {
			JPlag jplag = new JPlag(options);
			JPlagResult result = jplag.run();

			Long currentTime = System.currentTimeMillis();
			DuplicateCheckResponse.Builder builder = DuplicateCheckResponse.newBuilder();
			DuplicateCheckResponseValue.Builder comparisonBuilder = DuplicateCheckResponseValue.newBuilder();
			builder.setTimeStamp(Long.toString(currentTime));

			List<JPlagComparison> comparisons = result.getComparisons();
			for (int i = 0; i < comparisons.size(); i++) {
				JPlagComparison comparison = comparisons.get(i);
				DuplicateCheckResponseValue.Comparsion.Builder comparisonsBuilder = DuplicateCheckResponseValue.Comparsion.newBuilder();
				long firstID, secondID;
				try {
					firstID = Long.parseLong(comparison.getFirstSubmission().getName());
					secondID = Long.parseLong(comparison.getSecondSubmission().getName());
				} catch (NumberFormatException n) {
					// 遇到非数字则结束遍历
					break;
				}

				comparisonsBuilder
						.setSimilarity((int) (comparison.similarityOfSecond() * 100))
						.setUserId(firstID)
						.setAnotherUserId(secondID)
						.setHtmlFileName(String.format(htmlFileLayout, i));

				comparisonBuilder.addComparisions(comparisonsBuilder);
			}

			String outputPath = Paths
					.get(outputPathPrefix, Long.toString(request.getLabID()), Long.toString(currentTime))
					.toString();

			File output = new File(outputPath);
			Report report = new Report(output, options);
			report.writeResult(result);

			builder.setComparision(comparisonBuilder);
			responseObserver.onNext(builder.build());
			responseObserver.onCompleted();
		} catch (SubmissionException e) {
			responseObserver.onError(Status.DATA_LOSS.withDescription(e.getMessage()).asException());
		} catch (UnsupportedOperationException | RootDirectoryException e) {
			responseObserver.onError(Status.FAILED_PRECONDITION.withDescription(e.getMessage()).asException());
		} catch (IllegalStateException e) {
			responseObserver.onError(Status.ABORTED.withDescription(e.getMessage()).asException());
		} catch (ReportGenerationException e) {
			responseObserver.onError(Status.INTERNAL.withDescription(e.getMessage()).asException());
		} catch (Exception e) {
			responseObserver.onError(Status.UNKNOWN.withDescription(e.getMessage()).asException());
		}
	}

	@Override
	public void viewReport(ViewReportRequest request, StreamObserver<ViewReportResponse> responseObserver) {
		Path resultPath = Paths.get(outputPathPrefix, Long.toString(request.getLabId()), request.getTimeStamp(),
				request.getHtmlFileName());

		File resultFile = resultPath.toFile();
		if (!resultFile.exists()) {
			responseObserver.onError(Status.NOT_FOUND.withDescription("lab dir is not found").asException());
			return;
		}

		FileReader fileReader;
		StringBuilder sb = new StringBuilder();
		try {
			fileReader = new FileReader(resultFile);
			char[] buf = new char[1024];
			for (int ch = fileReader.read(buf); ch != -1; ch = fileReader.read(buf)) {
				if (ch != 1024) {
					sb.append(Arrays.copyOfRange(buf, 0, ch));
				} else {
					sb.append(buf);
				}
			}
		} catch (FileNotFoundException e) {
			responseObserver.onError(Status.NOT_FOUND.withDescription("lab dir is not found").asException());
			return;
		} catch (IOException e) {
			responseObserver.onError(Status.INTERNAL.withDescription("read html file failed").asException());
			return;
		}

		ViewReportResponse.Builder builder = ViewReportResponse.newBuilder();
		builder.setHtmlFileContent(sb.toString());
		responseObserver.onNext(builder.build());
		responseObserver.onCompleted();
	}

	@Override
	public void generateTestFilesForDuplicateCheck(plagiarismDetectionServerImplBase.GenerateTestFilesForDuplicateCheckRequest request, StreamObserver<Empty> responseObserver) {
		String fileName = null;
		language lan = request.getLan();
		if (lan == language.python3) {
			fileName = "1.py";
		} else if (lan == language.cpp) {
			fileName = "1.cpp";
		} else {
			fileName = "Solution.java";
		}

		for (int j = 1; j <= 2; j++) {
			try {
				Path dirPath = Paths.get(baseCodePath, "workspace-0", Integer.toString(j));
				if (!dirPath.toFile().exists()) {
					dirPath.toFile().mkdirs();
				}
				Path codePath = Paths.get(dirPath.toString(), fileName);
				FileWriter fileWriter = new FileWriter(codePath.toString());
				fileWriter.write(request.getCodeContent());
				fileWriter.close();
			} catch (IOException e) {
				responseObserver.onError(Status.INTERNAL.withDescription(e.getMessage()).asException());
				return;
			}
		}
		Empty.Builder builder = Empty.newBuilder();
		responseObserver.onNext(builder.build());
		responseObserver.onCompleted();
	}

	@Override
	public void removeTestFilesForDuplicateCheck(plagiarismDetectionServerImplBase.Empty request, StreamObserver<Empty> responseObserver) {
		Path dirPath = Paths.get(baseCodePath, "workspace-0");
		delete(dirPath);
		Empty.Builder builder = Empty.newBuilder();
		responseObserver.onNext(builder.build());
		responseObserver.onCompleted();
	}

	private static void delete(Path path) {
		if (!Files.exists(path)) {
			return;
		}

		try {
			if (!Files.isDirectory(path)) {
				Files.delete(path);
				return;
			}
			Files.list(path).forEach(f -> delete(f));
		} catch (IOException e) {
			e.printStackTrace();
		}
	}

	@Override
	public void generateTestHTMLFileForViewReport(plagiarismDetectionServerImplBase.GenerateTestHTMLFileForViewReportRequest request, StreamObserver<Empty> responseObserver) {
		Path dirPath = Paths.get(outputPathPrefix, "0", request.getTimeStamp());
		if (!dirPath.toFile().exists()) {
			dirPath.toFile().mkdirs();
		}
		Path codePath = Paths.get(dirPath.toString(), request.getHtmlFileName());
		FileWriter fileWriter = null;
		try {
			fileWriter = new FileWriter(codePath.toString());
			fileWriter.close();
		} catch (IOException e) {
			responseObserver.onError(Status.INTERNAL.withDescription(e.getMessage()).asException());
			return;
		}
		Empty.Builder builder = Empty.newBuilder();
		responseObserver.onNext(builder.build());
		responseObserver.onCompleted();
	}

	@Override
	public void removeTestHTMLFileForViewReport(Empty request, StreamObserver<Empty> responseObserver) {
		Path dirPath = Paths.get(outputPathPrefix, "0");
		delete(dirPath);
		Empty.Builder builder = Empty.newBuilder();
		responseObserver.onNext(builder.build());
		responseObserver.onCompleted();
	}
}