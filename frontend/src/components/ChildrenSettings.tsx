import { useState, useEffect, useMemo } from "react";
import { get } from "../api";
import { Child, ChildListResponse } from "../types";
import Card from "./ui/Card";
import ChildSelectorBar from "./ChildSelectorBar";
import AddChildForm from "./AddChildForm";
import ChildAccountSettings from "./ChildAccountSettings";

interface ChildrenSettingsProps {
  selectedChildName?: string;
  onChildSelect: (child: Child | null) => void;
}

export default function ChildrenSettings({
  selectedChildName,
  onChildSelect,
}: ChildrenSettingsProps) {
  const [childRefreshKey, setChildRefreshKey] = useState(0);
  const [children, setChildren] = useState<Child[]>([]);
  const [loading, setLoading] = useState(true);

  // Derive selected child from name prop
  const selectedChild = useMemo(() => {
    if (!selectedChildName || children.length === 0) return null;
    return children.find(
      (c) => c.first_name.toLowerCase() === selectedChildName.toLowerCase()
    ) ?? null;
  }, [selectedChildName, children]);

  // Redirect if child name is invalid
  useEffect(() => {
    if (selectedChildName && children.length > 0 && !selectedChild) {
      onChildSelect(null);
    }
  }, [selectedChildName, children, selectedChild, onChildSelect]);

  useEffect(() => {
    setLoading(true);
    get<ChildListResponse>("/children")
      .then((data) => {
        const list = data.children || [];
        setChildren(list);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [childRefreshKey]);

  const handleChildAdded = () => {
    setChildRefreshKey((k) => k + 1);
  };

  const handleChildUpdated = () => {
    setChildRefreshKey((k) => k + 1);
  };

  const handleChildDeleted = () => {
    onChildSelect(null);
    setChildRefreshKey((k) => k + 1);
  };

  return (
    <div className="space-y-4">
      <AddChildForm onChildAdded={handleChildAdded} />

      <ChildSelectorBar
        children={children}
        selectedChildId={selectedChild?.id ?? null}
        onSelectChild={onChildSelect}
        loading={loading}
      />

      {selectedChild ? (
        <ChildAccountSettings
          key={selectedChild.id}
          child={selectedChild}
          onUpdated={handleChildUpdated}
          onDeleted={handleChildDeleted}
        />
      ) : (
        !loading && children.length > 0 && (
          <Card padding="lg">
            <p className="text-bark-light text-center py-4">
              Select a child to manage their account settings.
            </p>
          </Card>
        )
      )}
    </div>
  );
}
