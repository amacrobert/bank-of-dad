import { useState, useEffect } from "react";
import { get } from "../api";
import { Child, ChildListResponse } from "../types";
import Card from "./ui/Card";
import ChildSelectorBar from "./ChildSelectorBar";
import AddChildForm from "./AddChildForm";
import ChildAccountSettings from "./ChildAccountSettings";

export default function ChildrenSettings() {
  const [childRefreshKey, setChildRefreshKey] = useState(0);
  const [children, setChildren] = useState<Child[]>([]);
  const [selectedChild, setSelectedChild] = useState<Child | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    get<ChildListResponse>("/children")
      .then((data) => {
        const list = data.children || [];
        setChildren(list);
        // If a child was selected, update with fresh data or deselect if removed
        if (selectedChild) {
          const updated = list.find((c) => c.id === selectedChild.id);
          if (updated) {
            setSelectedChild(updated);
          } else {
            setSelectedChild(null);
          }
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [childRefreshKey]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleChildAdded = () => {
    setChildRefreshKey((k) => k + 1);
  };

  const handleChildUpdated = () => {
    setChildRefreshKey((k) => k + 1);
  };

  const handleChildDeleted = () => {
    setSelectedChild(null);
    setChildRefreshKey((k) => k + 1);
  };

  return (
    <div className="space-y-4">
      <AddChildForm onChildAdded={handleChildAdded} />

      <ChildSelectorBar
        children={children}
        selectedChildId={selectedChild?.id ?? null}
        onSelectChild={setSelectedChild}
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
