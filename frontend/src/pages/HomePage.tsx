import { Shield, Coins, TrendingUp, Leaf } from "lucide-react";
import Card from "../components/ui/Card";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-cream overflow-hidden">
      {/* Decorative background blobs */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden" aria-hidden="true">
        <div className="absolute -top-32 -right-32 w-80 h-80 bg-sage/15 rounded-full blur-3xl" />
        <div className="absolute top-1/3 -left-24 w-64 h-64 bg-amber/10 rounded-full blur-3xl" />
        <div className="absolute bottom-20 right-10 w-48 h-48 bg-sage-light/20 rounded-full blur-3xl" />
      </div>

      <div className="relative max-w-lg mx-auto px-4 py-12 md:py-20">
        {/* Hero */}
        <div className="text-center mb-12 animate-fade-in-up">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-forest/10 rounded-2xl mb-6">
            <Leaf className="h-8 w-8 text-forest" aria-hidden="true" />
          </div>
          <h1 className="text-4xl md:text-5xl font-extrabold text-forest mb-4 leading-tight">
            Bank of Dad
          </h1>
          <p className="text-xl text-bark-light max-w-sm mx-auto leading-relaxed">
            Watch your family's savings grow.
          </p>
        </div>

        {/* Google Sign In */}
        <div className="text-center mb-12 animate-fade-in-up" style={{ animationDelay: "0.1s" }}>
          <a
            href="/api/auth/google/login"
            className="
              inline-flex items-center gap-3
              min-h-[48px] px-8 py-3
              bg-white rounded-xl border border-sand
              shadow-[0_2px_8px_rgba(61,46,31,0.08)]
              text-bark font-semibold text-base
              hover:bg-cream-dark hover:shadow-[0_4px_12px_rgba(61,46,31,0.12)]
              active:scale-[0.97]
              transition-all duration-200
            "
          >
            <svg className="h-5 w-5" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 01-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/>
              <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
              <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
              <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
            </svg>
            Sign in with Google
          </a>
        </div>

        {/* Value props */}
        <div className="space-y-4 mb-12">
          {[
            { icon: Shield, title: "Parents set the rules", desc: "Full control over accounts, allowances, and spending." },
            { icon: Coins, title: "Kids learn by doing", desc: "Real balances, real transactions, real financial literacy." },
            { icon: TrendingUp, title: "Watch savings grow", desc: "Compound interest shows the power of patience." },
          ].map((prop, i) => (
            <div key={prop.title} className="animate-fade-in-up" style={{ animationDelay: `${0.2 + i * 0.1}s` }}>
              <Card padding="md">
                <div className="flex items-start gap-4">
                  <div className="flex-shrink-0 w-10 h-10 bg-sage-light/40 rounded-xl flex items-center justify-center">
                    <prop.icon className="h-5 w-5 text-forest" aria-hidden="true" />
                  </div>
                  <div>
                    <h3 className="font-bold text-bark text-base mb-1">{prop.title}</h3>
                    <p className="text-bark-light text-sm leading-relaxed">{prop.desc}</p>
                  </div>
                </div>
              </Card>
            </div>
          ))}
        </div>

        {/* Footer */}
        <footer className="text-center text-sm text-bark-light/60 pb-8">
          <p>Bank of Dad</p>
        </footer>
      </div>
    </div>
  );
}
