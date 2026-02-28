import { useEffect, useState, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import { getSubscription, createCheckoutSession, createPortalSession, ApiRequestError } from "../api";
import { SubscriptionResponse } from "../types";
import Card from "./ui/Card";
import Button from "./ui/Button";
import LoadingSpinner from "./ui/LoadingSpinner";

export default function SubscriptionSettings() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [subscription, setSubscription] = useState<SubscriptionResponse | null>(null);
  const [error, setError] = useState("");
  const [checkoutLoading, setCheckoutLoading] = useState<string | null>(null);
  const [portalLoading, setPortalLoading] = useState(false);
  const [polling, setPolling] = useState(false);

  const fetchSubscription = useCallback(async () => {
    try {
      const data = await getSubscription();
      setSubscription(data);
      return data;
    } catch {
      setError("Failed to load subscription status.");
      return null;
    }
  }, []);

  useEffect(() => {
    fetchSubscription().then(() => setLoading(false));
  }, [fetchSubscription]);

  // T023: Handle checkout return with polling
  useEffect(() => {
    if (searchParams.get("success") !== "true") return;

    setPolling(true);
    let attempts = 0;
    const maxAttempts = 15;

    const pollInterval = setInterval(async () => {
      attempts++;
      const data = await fetchSubscription();
      if (data && data.account_type === "plus") {
        clearInterval(pollInterval);
        setPolling(false);
        setSearchParams({}, { replace: true });
      } else if (attempts >= maxAttempts) {
        clearInterval(pollInterval);
        setPolling(false);
        setSearchParams({}, { replace: true });
      }
    }, 2000);

    return () => clearInterval(pollInterval);
  }, [searchParams, fetchSubscription, setSearchParams]);

  const handleUpgrade = async (lookupKey: string) => {
    setCheckoutLoading(lookupKey);
    setError("");
    try {
      const { checkout_url } = await createCheckoutSession(lookupKey);
      window.location.href = checkout_url;
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.error || "Failed to start checkout.");
      } else {
        setError("Failed to start checkout.");
      }
      setCheckoutLoading(null);
    }
  };

  const handleManageSubscription = async () => {
    setPortalLoading(true);
    setError("");
    try {
      const { portal_url } = await createPortalSession();
      window.location.href = portal_url;
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.error || "Failed to open subscription portal.");
      } else {
        setError("Failed to open subscription portal.");
      }
      setPortalLoading(false);
    }
  };

  if (loading) {
    return <LoadingSpinner message="Loading subscription..." />;
  }

  if (polling) {
    return (
      <Card>
        <div className="flex flex-col items-center gap-4 py-8">
          <LoadingSpinner message="Processing your subscription..." />
          <p className="text-sm text-bark-light">This may take a few moments.</p>
        </div>
      </Card>
    );
  }

  if (!subscription) {
    return (
      <Card>
        <p className="text-terracotta">{error || "Unable to load subscription information."}</p>
      </Card>
    );
  }

  const isPlus = subscription.account_type === "plus";
  const isCancelling = isPlus && subscription.cancel_at_period_end;

  return (
    <div className="space-y-4">
      <Card>
        <h2 className="text-lg font-bold text-bark mb-4">Subscription</h2>

        {/* Current plan display */}
        <div className="flex items-center gap-3 mb-6">
          <span className="text-sm font-medium text-bark-light">Current Plan</span>
          <span
            className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-bold ${
              isPlus
                ? "bg-forest/10 text-forest"
                : "bg-sand text-bark-light"
            }`}
          >
            {isPlus ? "Plus" : "Free"}
          </span>
        </div>

        {/* Plus subscriber details */}
        {isPlus && (
          <div className="space-y-3 mb-6">
            {subscription.subscription_status && (
              <div className="flex justify-between text-sm">
                <span className="text-bark-light">Status</span>
                <span className="font-medium text-bark capitalize">
                  {subscription.subscription_status === "active" && !isCancelling && "Active"}
                  {subscription.subscription_status === "active" && isCancelling && "Cancelling"}
                  {subscription.subscription_status === "past_due" && "Past Due"}
                  {subscription.subscription_status !== "active" && subscription.subscription_status !== "past_due" && subscription.subscription_status}
                </span>
              </div>
            )}

            {subscription.current_period_end && (
              <div className="flex justify-between text-sm">
                <span className="text-bark-light">
                  {isCancelling ? "Access Until" : "Next Renewal"}
                </span>
                <span className="font-medium text-bark">
                  {new Date(subscription.current_period_end).toLocaleDateString(undefined, {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })}
                </span>
              </div>
            )}

            {isCancelling && (
              <div className="bg-amber-50 border border-amber-200 rounded-xl px-4 py-3 text-sm text-amber-800">
                Your subscription will end on{" "}
                {subscription.current_period_end &&
                  new Date(subscription.current_period_end).toLocaleDateString(undefined, {
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })}
                . You'll retain Plus features until then.
              </div>
            )}

            {subscription.subscription_status === "past_due" && (
              <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl px-4 py-3 text-sm text-terracotta">
                Your payment is past due. Please update your payment method to keep Plus features.
              </div>
            )}
          </div>
        )}

        {/* Actions */}
        {isPlus ? (
          <Button
            onClick={handleManageSubscription}
            loading={portalLoading}
            disabled={portalLoading}
          >
            Manage Subscription
          </Button>
        ) : (
          <div className="space-y-3">
            <p className="text-sm text-bark-light mb-4">
              Upgrade to Plus to unlock premium features for your family.
            </p>
            <div className="flex flex-col sm:flex-row gap-3">
              <Button
                onClick={() => handleUpgrade("plus_monthly")}
                loading={checkoutLoading === "plus_monthly"}
                disabled={checkoutLoading !== null}
              >
                Monthly — $1/mo
              </Button>
              <Button
                onClick={() => handleUpgrade("plus_annual")}
                loading={checkoutLoading === "plus_annual"}
                disabled={checkoutLoading !== null}
                variant="secondary"
              >
                Annual — $10/yr
              </Button>
            </div>
          </div>
        )}

        {error && (
          <p className="mt-4 text-sm font-medium text-terracotta">{error}</p>
        )}
      </Card>
    </div>
  );
}
