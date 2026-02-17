if (!process.env.NEXT_PUBLIC_SERVER_URL) {
  throw new Error("NEXT_PUBLIC_SERVER_URL environment variable is not set, and is required.");
}

export const SERVER_URL = `${process.env.NEXT_PUBLIC_SERVER_URL}`;
