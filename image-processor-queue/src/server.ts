import "dotenv/config";
import app from "./app.js";
import prisma from "./prisma.js";

const PORT = process.env.PORT;

async function startServer() {
  try {
    await prisma.$connect();
    console.log("✅ Successfully connected to the Postgres Database!");

    app.listen(PORT, () => {
      console.log(`🚀 Server is running on http://localhost:${PORT}`);
    });
  } catch (error) {
    console.error(
      "❌ Failed to start the server or connect to the database:",
      error,
    );
    process.exit(1);
  }
}

startServer();
