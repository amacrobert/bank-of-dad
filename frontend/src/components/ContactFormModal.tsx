import { useState } from "react";
import Modal from "./ui/Modal";
import Button from "./ui/Button";
import { submitContact } from "../api";

interface ContactFormModalProps {
  open: boolean;
  onClose: () => void;
}

const MAX_LENGTH = 1000;

export default function ContactFormModal({ open, onClose }: ContactFormModalProps) {
  const [body, setBody] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const handleClose = () => {
    setBody("");
    setError("");
    setSuccess(false);
    setSubmitting(false);
    onClose();
  };

  const handleSubmit = async () => {
    const trimmed = body.trim();
    if (!trimmed) {
      setError("Please enter a message.");
      return;
    }
    setError("");
    setSubmitting(true);
    try {
      await submitContact(trimmed);
      setSuccess(true);
      setTimeout(handleClose, 1500);
    } catch {
      setError("Failed to send message. Please try again.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal open={open} onClose={handleClose}>
      <h2 className="text-lg font-bold text-bark mb-1">Contact Us</h2>
      <p className="text-sm text-bark-light mb-4">
        Send us a message — feedback, feature requests, bug reports, or anything else.
      </p>

      {success ? (
        <p className="text-sm text-forest font-medium py-4 text-center">
          Message sent! Thank you.
        </p>
      ) : (
        <>
          <textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            maxLength={MAX_LENGTH}
            rows={5}
            placeholder="What's on your mind?"
            className="w-full rounded-xl border border-sand bg-white px-4 py-3 text-sm text-bark placeholder:text-bark-light/50 focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest resize-none"
          />
          <div className="flex items-center justify-between mt-1 mb-4">
            {error ? (
              <p className="text-sm text-terracotta">{error}</p>
            ) : (
              <span />
            )}
            <span className="text-xs text-bark-light/50">
              {body.length}/{MAX_LENGTH}
            </span>
          </div>

          <div className="flex gap-3 justify-end">
            <Button variant="secondary" onClick={handleClose}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? "Sending..." : "Send"}
            </Button>
          </div>
        </>
      )}
    </Modal>
  );
}
