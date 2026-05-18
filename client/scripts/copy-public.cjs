const fs = require("fs").promises;
const path = require("path");

async function copyDir(src, dest) {
  await fs.mkdir(dest, { recursive: true });
  const entries = await fs.readdir(src, { withFileTypes: true });
  for (const entry of entries) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);
    if (entry.isDirectory()) {
      await copyDir(srcPath, destPath);
    } else {
      await fs.copyFile(srcPath, destPath);
    }
  }
}

async function main() {
  const repoRoot = path.resolve(__dirname, "..", "..");
  const src = path.join(repoRoot, "client.old", "public");
  const dest = path.join(__dirname, "..", "public");

  try {
    await copyDir(src, dest);
    console.log(`Copied public assets from ${src} to ${dest}`);
  } catch (err) {
    console.error("Error copying public assets:", err);
    process.exitCode = 1;
  }
}

main();
