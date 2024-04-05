import java.io.IOException;
import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

import io.grpc.Server;
import io.grpc.ServerBuilder;

public class plagiarismDetectionServer {

    private Server server;
    private static final int port = 8086;

    private static final Logger logger = Logger.getLogger(plagiarismDetectionServer.class.getName());

    private void start() throws IOException {
        server = ServerBuilder
                .forPort(port)
                .addService(new plagiarismDetectionService())
                .build()
                .start();
        logger.info("Server has started");

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            try {
                plagiarismDetectionServer.this.stop();
            } catch (InterruptedException e) {
                e.printStackTrace();
            }
            logger.info("Server shutdown");
        }));
    }

    private void stop() throws InterruptedException {
        if (null != server) {
            logger.info("Trying to shutdown GRPC Server");
            server.shutdown().awaitTermination(10, TimeUnit.SECONDS);
        }
    }

    private void waitInsteadShutdown() throws InterruptedException {
        if (null != server) {
            server.awaitTermination();
        }
    }

    public static void main(String[] args) throws IOException, InterruptedException {
        final plagiarismDetectionServer server = new plagiarismDetectionServer();
        server.start();
        server.waitInsteadShutdown();
    }
}