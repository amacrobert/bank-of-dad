import { useEffect, useState } from "react";
import { get } from "../api";
import { ChildListResponse, Child } from "../types";
import BalanceDisplay from "./BalanceDisplay";

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
    return (
      <div className="child-list-empty">
        <p>No children added yet. Add your first child above!</p>
        <p className="hint">Once you add children, you can deposit money into their accounts.</p>
      </div>
    );
  }

  return (
    <div className="child-list">
      <h3>Children</h3>
      <ul>
        {children.map((child) => (
          <li key={child.id} className="child-item">
            <div className="child-info">
              <span className="child-name">{child.first_name}</span>
              {child.is_locked && <span className="child-locked"> (locked)</span>}
            </div>
            <div className="child-balance">
              <BalanceDisplay balanceCents={child.balance_cents} />
            </div>
            <div className="child-actions">
              {onSelectChild && (
                <button
                  onClick={() => onSelectChild(child)}
                  className="btn-manage"
                >
                  Manage
                </button>
              )}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
