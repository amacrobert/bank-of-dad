import { useOutletContext } from "react-router-dom";
import { ParentUser, ChildUser, AuthUser } from "../types";

interface ParentOutletContext {
  user: ParentUser;
  setUser: React.Dispatch<React.SetStateAction<AuthUser | null>>;
}

interface ChildOutletContext {
  user: ChildUser;
  setUser: React.Dispatch<React.SetStateAction<AuthUser | null>>;
}

export function useParentUser(): ParentUser {
  const { user } = useOutletContext<ParentOutletContext>();
  return user;
}

export function useChildUser(): ChildUser {
  const { user } = useOutletContext<ChildOutletContext>();
  return user;
}

export function useSetUser(): React.Dispatch<React.SetStateAction<AuthUser | null>> {
  const { setUser } = useOutletContext<ChildOutletContext>();
  return setUser;
}
