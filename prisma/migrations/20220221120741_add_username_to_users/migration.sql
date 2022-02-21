/*
  Warnings:

  - A unique constraint covering the columns `[normalized_username]` on the table `users` will be added. If there are existing duplicate values, this will fail.
  - Added the required column `normalized_username` to the `users` table without a default value. This is not possible if the table is not empty.
  - Added the required column `username` to the `users` table without a default value. This is not possible if the table is not empty.

*/
-- DropIndex
DROP INDEX "users_email_key";

-- AlterTable
ALTER TABLE "users" ADD COLUMN     "normalized_username" TEXT NOT NULL,
ADD COLUMN     "username" TEXT NOT NULL;

-- CreateIndex
CREATE UNIQUE INDEX "users_normalized_username_key" ON "users"("normalized_username");

-- CreateIndex
CREATE INDEX "users_normalized_username_idx" ON "users"("normalized_username");
