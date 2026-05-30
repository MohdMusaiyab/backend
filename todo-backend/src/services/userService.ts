import { prisma } from "../lib/prisma.js";
import type { UpdateUserInput } from "../schemas/user.schema.js";

export const getUserProfileService = async (userId: string) => {
  const user = await prisma.user.findUnique({
    where: { id: userId },
    select: {
      id: true,
      email: true,
      todo: {
        orderBy: { createdAt: "desc" },
        take: 3,
      },
    },
  });

  if (!user) {
    throw new Error("User not found");
  }

  return user;
};

export const updateUserService = async (
  userId: string,
  data: UpdateUserInput,
) => {
  const user = await prisma.user.findUnique({ where: { id: userId } });
  if (!user) throw new Error("User not found");

  if (data.email && data.email !== user.email) {
    const existing = await prisma.user.findUnique({
      where: { email: data.email },
    });
    if (existing) throw new Error("Email is already in use");
  }

  const updateData: any = {};
  if (data.email !== undefined) updateData.email = data.email;

  const updatedUser = await prisma.user.update({
    where: { id: userId },
    data: updateData,
    select: {
      id: true,
      email: true,
    },
  });

  return updatedUser;
};

export const deleteUserService = async (userId: string) => {
  await prisma.user.delete({
    where: { id: userId },
  });
};
