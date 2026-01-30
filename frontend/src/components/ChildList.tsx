import { useEffect, useState } from "react";
import { get } from "../api";
import { ChildListResponse, Child } from "../types";

interface ChildListProps {
  refreshKey: number;
  onSelectChild?: (child: Child) => void;
}

export default function ChildList({ refreshKey, onSelectChild }: ChildListProps) {
  const [children, setChildren] = useState<Child[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    get<ChildListResponse>("/children")
      .then((data) => {
        setChildren(data.children || []);
        setLoading(false);
      })
      .catch(() => {
        setLoading(false);
      });
  }, [refreshKey]);

  if (loading) {
    return <p>Loading children...</p>;
  }

  if (children.length === 0) {
    return <p>No children added yet. Add your first child above!</p>;
  }

  return (
    <div className="child-list">
      <h3>Children</h3>
      <ul>
        {children.map((child) => (
          <li key={child.id} className="child-item">
            <span className="child-name">{child.first_name}</span>
            {child.is_locked && <span className="child-locked"> (locked)</span>}
            <span className="child-date">
              Added {new Date(child.created_at).toLocaleDateString()}
            </span>
            {onSelectChild && (
              <button
                onClick={() => onSelectChild(child)}
                className="btn-manage"
              >
                Manage
              </button>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
