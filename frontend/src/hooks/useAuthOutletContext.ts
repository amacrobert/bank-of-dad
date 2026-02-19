import { useOutletContext } from "react-router-dom";
import { ParentUser, ChildUser } from "../types";

interface ParentOutletContext {
  user: ParentUser;
}

interface ChildOutletContext {
  user: ChildUser;
}

export function useParentUser(): ParentUser {
  const { user } = useOutletContext<ParentOutletContext>();
  return user;
}

export function useChildUser(): ChildUser {
  const { user } = useOutletContext<ChildOutletContext>();
  return user;
}
