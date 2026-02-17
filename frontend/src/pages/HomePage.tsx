import { useRef, useState, useEffect, ReactNode } from "react";
import {
  Shield,
  Coins,
  TrendingUp,
  Leaf,
  Sparkles,
  Users,
  Calendar,
  Clock,
  Lock,
  CheckCircle,
  UserPlus,
  ArrowDownCircle,
} from "lucide-react";
import Card from "../components/ui/Card";

// ---------------------------------------------------------------------------
// useInView — fires once when element enters viewport
// ---------------------------------------------------------------------------
function useInView(threshold = 0.15) {
  const ref = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setVisible(true);
          observer.unobserve(el);
        }
      },
      { threshold },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [threshold]);

  return { ref, visible };
}

// ---------------------------------------------------------------------------
// AnimatedSection — fades/slides children in on scroll
// ---------------------------------------------------------------------------
function AnimatedSection({
  children,
  className = "",
  direction = "up",
}: {
  children: ReactNode;
  className?: string;
  direction?: "up" | "left" | "right";
}) {
  const { ref, visible } = useInView();
  const animClass =
    direction === "left"
      ? "animate-fade-in-left"
      : direction === "right"
        ? "animate-fade-in-right"
        : "animate-fade-in-up";
  return (
    <div
      ref={ref}
      className={`${visible ? animClass : "opacity-0"} ${className}`}
    >
      {children}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Google Sign-In Button
// ---------------------------------------------------------------------------
const googleLoginUrl = `${import.meta.env.VITE_API_URL || ""}/api/auth/google/login`;

const GoogleSvg = () => (
  <svg className="h-5 w-5 shrink-0" viewBox="0 0 24 24" aria-hidden="true">
    <path
      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 01-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
      fill="#4285F4"
    />
    <path
      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      fill="#34A853"
    />
    <path
      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
      fill="#FBBC05"
    />
    <path
      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      fill="#EA4335"
    />
  </svg>
);

function GoogleSignInButton({ size = "default" }: { size?: "compact" | "default" | "large" }) {
  const base =
    "inline-flex items-center justify-center gap-2.5 bg-white rounded-xl border border-sand shadow-[0_2px_8px_rgba(61,46,31,0.08)] text-bark font-semibold hover:bg-cream-dark hover:shadow-[0_4px_12px_rgba(61,46,31,0.12)] active:scale-[0.97] transition-all duration-200";

  const sizeClass =
    size === "compact"
      ? "px-4 py-2 text-sm"
      : size === "large"
        ? "px-10 py-4 text-lg"
        : "px-8 py-3 text-base";

  return (
    <a href={googleLoginUrl} className={`${base} ${sizeClass}`}>
      <GoogleSvg />
      Sign in with Google
    </a>
  );
}

function GoogleSignInButtonDark({ size = "large" }: { size?: "default" | "large" }) {
  const base =
    "inline-flex items-center justify-center gap-2.5 bg-white rounded-xl border border-white/20 shadow-[0_4px_16px_rgba(0,0,0,0.2)] text-bark font-semibold hover:bg-cream hover:shadow-[0_6px_24px_rgba(0,0,0,0.25)] active:scale-[0.97] transition-all duration-200";

  const sizeClass = size === "large" ? "px-10 py-4 text-lg" : "px-8 py-3 text-base";

  return (
    <a href={googleLoginUrl} className={`${base} ${sizeClass}`}>
      <GoogleSvg />
      Sign in with Google
    </a>
  );
}

// ---------------------------------------------------------------------------
// Mock Dashboard Components (decorative only)
// ---------------------------------------------------------------------------
function MockChildDashboard() {
  return (
    <Card className="max-w-xs mx-auto" padding="md">
      <p className="text-sm font-semibold text-bark-light mb-1">Welcome, Emma!</p>
      <p className="text-3xl font-extrabold text-forest mb-2">$124.50</p>
      <span className="inline-flex items-center gap-1 text-xs font-medium bg-sage-light/30 text-forest px-2 py-0.5 rounded-full mb-3">
        <TrendingUp className="h-3 w-3" /> 2% monthly interest
      </span>
      <div className="space-y-2 border-t border-sand pt-3">
        {[
          { label: "Weekly allowance", amount: "+$5.00", color: "text-forest" },
          { label: "Toy shop", amount: "-$12.00", color: "text-terracotta" },
          { label: "Interest earned", amount: "+$2.49", color: "text-forest" },
        ].map((tx) => (
          <div key={tx.label} className="flex justify-between text-sm">
            <span className="text-bark-light">{tx.label}</span>
            <span className={`font-semibold ${tx.color}`}>{tx.amount}</span>
          </div>
        ))}
      </div>
    </Card>
  );
}

function MockParentDashboard() {
  return (
    <Card className="max-w-xs mx-auto" padding="md">
      <p className="text-sm font-semibold text-bark-light mb-3">Your Family Bank</p>
      <div className="space-y-3">
        {[
          { name: "Emma", balance: "$124.50", color: "bg-sage-light" },
          { name: "Jack", balance: "$87.25", color: "bg-amber-light" },
        ].map((child) => (
          <div key={child.name} className="flex items-center gap-3">
            <div
              className={`w-8 h-8 ${child.color} rounded-full flex items-center justify-center text-xs font-bold text-forest`}
            >
              {child.name[0]}
            </div>
            <div className="flex-1">
              <p className="text-sm font-semibold text-bark">{child.name}</p>
            </div>
            <p className="text-sm font-bold text-forest">{child.balance}</p>
          </div>
        ))}
      </div>
      <div className="mt-3 pt-3 border-t border-sand text-xs text-bark-light">
        Next allowance: Tomorrow
      </div>
    </Card>
  );
}

// ===========================================================================
// HomePage
// ===========================================================================
export default function HomePage() {
  return (
    <div className="min-h-screen bg-cream">
      {/* ----------------------------------------------------------------- */}
      {/* Sticky Header                                                      */}
      {/* ----------------------------------------------------------------- */}
      <header className="sticky top-0 z-50 bg-cream/90 backdrop-blur-md border-b border-sand/50">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Leaf className="h-6 w-6 text-forest" aria-hidden="true" />
            <span className="text-lg font-bold text-forest">Bank of Dad</span>
          </div>
          <GoogleSignInButton size="compact" />
        </div>
      </header>

      {/* ----------------------------------------------------------------- */}
      {/* Hero Section                                                       */}
      {/* ----------------------------------------------------------------- */}
      <section className="max-w-6xl mx-auto px-4 sm:px-6 pt-16 pb-20 md:pt-24 md:pb-28">
        <div className="grid md:grid-cols-2 gap-12 items-center">
          {/* Text column */}
          <div className="animate-fade-in-up">
            <div className="inline-flex items-center gap-1.5 bg-amber-light/30 text-bark font-medium text-sm px-3 py-1 rounded-full mb-6">
              <Sparkles className="h-4 w-4 text-amber" aria-hidden="true" />
              Family Finance Made Simple
            </div>
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold text-forest leading-tight mb-6">
              Teach your kids the value of money
            </h1>
            <p className="text-lg text-bark-light leading-relaxed mb-8 max-w-lg">
              A simple family banking app where parents manage allowances,
              interest, and transactions — and kids learn real financial skills
              by watching their own savings grow.
            </p>
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
              <GoogleSignInButton />
              <span className="text-sm text-bark-light">
                Free to use &middot; No credit card needed
              </span>
            </div>
          </div>

          {/* Decorative app preview — hidden on mobile */}
          <div className="hidden md:block" aria-hidden="true">
            <div className="animate-float rotate-2">
              <MockChildDashboard />
            </div>
          </div>
        </div>
      </section>

      {/* ----------------------------------------------------------------- */}
      {/* Trust Strip                                                        */}
      {/* ----------------------------------------------------------------- */}
      <section className="bg-white/60 border-y border-sand/40">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 py-6">
          <div className="flex flex-col sm:flex-row items-center justify-center gap-6 sm:gap-12">
            {[
              { icon: Shield, text: "100% Free" },
              { icon: Lock, text: "Private & Secure" },
              { icon: Clock, text: "2 minutes to set up" },
            ].map(({ icon: Icon, text }) => (
              <div key={text} className="flex items-center gap-2 text-bark-light">
                <Icon className="h-5 w-5 text-forest" aria-hidden="true" />
                <span className="text-sm font-semibold">{text}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ----------------------------------------------------------------- */}
      {/* How It Works                                                       */}
      {/* ----------------------------------------------------------------- */}
      <section className="max-w-5xl mx-auto px-4 sm:px-6 py-20">
        <AnimatedSection className="text-center mb-14">
          <h2 className="text-3xl sm:text-4xl font-extrabold text-forest mb-3">
            How it works
          </h2>
          <p className="text-bark-light text-lg max-w-lg mx-auto">
            Get started in three simple steps
          </p>
        </AnimatedSection>

        <div className="grid md:grid-cols-3 gap-10">
          {[
            {
              step: 1,
              icon: Users,
              title: "Create your family bank",
              desc: "Sign in with Google and set up your family in seconds. No forms, no fuss.",
            },
            {
              step: 2,
              icon: UserPlus,
              title: "Add your kids",
              desc: "Create an account for each child with their own balance, allowance, and interest rate.",
            },
            {
              step: 3,
              icon: TrendingUp,
              title: "Watch them learn",
              desc: "Kids see their balances grow through allowances and compound interest — real lessons in real time.",
            },
          ].map(({ step, icon: Icon, title, desc }) => (
            <AnimatedSection key={step} className="text-center">
              <div className="relative inline-flex mb-4">
                <div className="w-14 h-14 bg-sage-light/30 rounded-2xl flex items-center justify-center">
                  <Icon className="h-7 w-7 text-forest" aria-hidden="true" />
                </div>
                <div className="absolute -top-2 -right-2 w-7 h-7 bg-forest text-white rounded-full flex items-center justify-center text-sm font-bold">
                  {step}
                </div>
              </div>
              <h3 className="text-xl font-bold text-bark mb-2">{title}</h3>
              <p className="text-bark-light leading-relaxed">{desc}</p>
            </AnimatedSection>
          ))}
        </div>
      </section>

      {/* ----------------------------------------------------------------- */}
      {/* Features Grid                                                      */}
      {/* ----------------------------------------------------------------- */}
      <section className="bg-white">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-20">
          <AnimatedSection className="text-center mb-14">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-forest mb-3">
              Everything you need
            </h2>
            <p className="text-bark-light text-lg max-w-lg mx-auto">
              Simple tools that make teaching money skills fun
            </p>
          </AnimatedSection>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[
              {
                icon: Shield,
                title: "Parental control",
                desc: "You control every deposit, withdrawal, and setting. Kids can view but parents run the bank.",
                tint: "bg-sage-light/30",
              },
              {
                icon: Calendar,
                title: "Automatic allowances",
                desc: "Set weekly or monthly allowances that deposit automatically. Set it and forget it.",
                tint: "bg-amber-light/30",
              },
              {
                icon: TrendingUp,
                title: "Compound interest",
                desc: "Add monthly interest to teach the power of saving. Watch those balances climb.",
                tint: "bg-sage-light/30",
              },
              {
                icon: ArrowDownCircle,
                title: "Deposits & withdrawals",
                desc: "Record chore payments, birthday money, or spending. Every transaction tells a story.",
                tint: "bg-amber-light/30",
              },
              {
                icon: Coins,
                title: "Upcoming payments",
                desc: "See when the next allowance or interest payment is due at a glance.",
                tint: "bg-sage-light/30",
              },
              {
                icon: Users,
                title: "Multiple children",
                desc: "Manage accounts for all your kids in one place, each with their own settings.",
                tint: "bg-amber-light/30",
              },
            ].map(({ icon: Icon, title, desc, tint }) => (
              <AnimatedSection key={title}>
                <Card padding="lg" className="h-full">
                  <div className={`inline-flex items-center justify-center w-11 h-11 ${tint} rounded-xl mb-4`}>
                    <Icon className="h-5 w-5 text-forest" aria-hidden="true" />
                  </div>
                  <h3 className="text-lg font-bold text-bark mb-2">{title}</h3>
                  <p className="text-bark-light text-sm leading-relaxed">{desc}</p>
                </Card>
              </AnimatedSection>
            ))}
          </div>
        </div>
      </section>

      {/* ----------------------------------------------------------------- */}
      {/* App Experience Preview                                             */}
      {/* ----------------------------------------------------------------- */}
      <section className="max-w-6xl mx-auto px-4 sm:px-6 py-20">
        <AnimatedSection className="text-center mb-14">
          <h2 className="text-3xl sm:text-4xl font-extrabold text-forest mb-3">
            Built for the whole family
          </h2>
          <p className="text-bark-light text-lg max-w-lg mx-auto">
            Separate views for parents and kids, each designed for what they need
          </p>
        </AnimatedSection>

        {/* Parent view */}
        <div className="grid md:grid-cols-2 gap-12 items-center mb-20">
          <AnimatedSection direction="left">
            <div className="inline-flex items-center gap-1.5 bg-sage-light/30 text-forest font-medium text-sm px-3 py-1 rounded-full mb-4">
              <Shield className="h-4 w-4" aria-hidden="true" />
              Parent View
            </div>
            <h3 className="text-2xl font-bold text-bark mb-4">
              Your family dashboard
            </h3>
            <ul className="space-y-3">
              {[
                "See all your children's balances at a glance",
                "Make deposits and withdrawals in seconds",
                "Set up automatic allowances and interest",
                "View full transaction history for each child",
              ].map((item) => (
                <li key={item} className="flex items-start gap-2.5">
                  <CheckCircle className="h-5 w-5 text-forest shrink-0 mt-0.5" aria-hidden="true" />
                  <span className="text-bark-light leading-relaxed">{item}</span>
                </li>
              ))}
            </ul>
          </AnimatedSection>
          <AnimatedSection direction="right" className="flex justify-center">
            <div aria-hidden="true">
              <MockParentDashboard />
            </div>
          </AnimatedSection>
        </div>

        {/* Child view */}
        <div className="grid md:grid-cols-2 gap-12 items-center">
          <AnimatedSection direction="left" className="md:order-2">
            <div className="inline-flex items-center gap-1.5 bg-amber-light/30 text-bark font-medium text-sm px-3 py-1 rounded-full mb-4">
              <Sparkles className="h-4 w-4 text-amber" aria-hidden="true" />
              Child View
            </div>
            <h3 className="text-2xl font-bold text-bark mb-4">
              Their own banking experience
            </h3>
            <ul className="space-y-3">
              {[
                "See their balance and recent transactions",
                "Watch savings grow with compound interest",
                "Learn how money works by doing, not just reading",
                "Simple, kid-friendly interface they can use on their own",
              ].map((item) => (
                <li key={item} className="flex items-start gap-2.5">
                  <CheckCircle className="h-5 w-5 text-amber shrink-0 mt-0.5" aria-hidden="true" />
                  <span className="text-bark-light leading-relaxed">{item}</span>
                </li>
              ))}
            </ul>
          </AnimatedSection>
          <AnimatedSection direction="right" className="md:order-1 flex justify-center">
            <div aria-hidden="true">
              <MockChildDashboard />
            </div>
          </AnimatedSection>
        </div>
      </section>

      {/* ----------------------------------------------------------------- */}
      {/* Final CTA                                                          */}
      {/* ----------------------------------------------------------------- */}
      <section className="bg-forest">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 py-20 text-center">
          <AnimatedSection>
            <Leaf className="h-10 w-10 text-sage-light mx-auto mb-6" aria-hidden="true" />
            <h2 className="text-3xl sm:text-4xl font-extrabold text-white mb-4">
              Start your family bank today
            </h2>
            <p className="text-sage-light text-lg mb-8 max-w-md mx-auto leading-relaxed">
              It's free, takes two minutes, and your kids will thank you
              (eventually).
            </p>
            <GoogleSignInButtonDark />
          </AnimatedSection>
        </div>
      </section>

      {/* ----------------------------------------------------------------- */}
      {/* Footer                                                             */}
      {/* ----------------------------------------------------------------- */}
      <footer className="bg-bark">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 py-8 flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Leaf className="h-5 w-5 text-sage" aria-hidden="true" />
            <span className="text-sm font-bold text-white">Bank of Dad</span>
          </div>
          <p className="text-sm text-sage-light/70">
            Teaching kids the value of money, one allowance at a time.
          </p>
        </div>
      </footer>
    </div>
  );
}
