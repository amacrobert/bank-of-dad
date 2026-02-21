export interface ThemeDefinition {
  label: string;
  slug: string;
  colors: {
    forest: string;
    forestLight: string;
    cream: string;
    creamDark: string;
  };
  backgroundSvg: string;
}

const saplingLeavesSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#2D5A3D' opacity='0.07'><path d='M0,-8 Q-4,1 0,8 Q4,1 0,-8Z' transform='translate(14,14) rotate(25)'/><path d='M0,-5 Q-3,0.5 0,5 Q3,0.5 0,-5Z' transform='translate(58,18) rotate(-20)'/><path d='M0,-9 Q-4.5,1 0,9 Q4.5,1 0,-9Z' transform='translate(33,42) rotate(40)'/><path d='M0,-4 Q-2.5,0.5 0,4 Q2.5,0.5 0,-4Z' transform='translate(72,56) rotate(-35)'/><path d='M0,-6 Q-3.5,0.5 0,6 Q3.5,0.5 0,-6Z' transform='translate(10,62) rotate(15)'/><path d='M0,-8 Q-4.5,1 0,8 Q4.5,1 0,-8Z' transform='translate(50,74) rotate(-10)'/><path d='M0,-4 Q-2,0.5 0,4 Q2,0.5 0,-4Z' transform='translate(75,14) rotate(50)'/><path d='M0,-5.5 Q-3,0.5 0,5.5 Q3,0.5 0,-5.5Z' transform='translate(38,8) rotate(-40)'/></g></svg>`)}")`;

const piggyBankCoinsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#8B4560' fill-opacity='0.3' stroke='#8B4560' opacity='0.08' stroke-width='1.5'><circle cx='14' cy='12' r='6'/><circle cx='58' cy='8' r='4.5'/><circle cx='38' cy='34' r='7'/><circle cx='73' cy='28' r='5'/><circle cx='10' cy='56' r='5.5'/><circle cx='55' cy='60' r='6'/><circle cx='28' cy='74' r='4'/><circle cx='74' cy='70' r='5'/></g></svg>`)}")`;

const sparkleStarsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#5B4BA0' opacity='0.06'><path d='M10 10l2 4 4 1-3 3 1 4-4-2-4 2 1-4-3-3 4-1z'/><path d='M55 8l1 3 3 0-2 2 1 3-3-1-3 1 1-3-2-2 3 0z'/><path d='M35 35l2 4 4 1-3 3 1 4-4-2-4 2 1-4-3-3 4-1z'/><path d='M70 40l1 3 3 0-2 2 1 3-3-1-3 1 1-3-2-2 3 0z'/><path d='M15 60l1 3 3 0-2 2 1 3-3-1-3 1 1-3-2-2 3 0z'/><path d='M60 68l2 4 4 1-3 3 1 4-4-2-4 2 1-4-3-3 4-1z'/></g></svg>`)}")`;

export const THEMES: Record<string, ThemeDefinition> = {
  sapling: {
    label: "Sapling",
    slug: "sapling",
    colors: {
      forest: "#2D5A3D",
      forestLight: "#3A7A52",
      cream: "#FDF6EC",
      creamDark: "#F5EBD9",
    },
    backgroundSvg: saplingLeavesSvg,
  },
  piggybank: {
    label: "Piggy Bank",
    slug: "piggybank",
    colors: {
      forest: "#8B4560",
      forestLight: "#A55A78",
      cream: "#FDF0F4",
      creamDark: "#F5E0E8",
    },
    backgroundSvg: piggyBankCoinsSvg,
  },
  sparkle: {
    label: "Sparkle",
    slug: "sparkle",
    colors: {
      forest: "#5B4BA0",
      forestLight: "#7366B8",
      cream: "#F3F0FD",
      creamDark: "#E8E3F5",
    },
    backgroundSvg: sparkleStarsSvg,
  },
};

export const THEME_SLUGS = Object.keys(THEMES);

export function getTheme(slug: string | null | undefined): ThemeDefinition {
  if (slug && THEMES[slug]) {
    return THEMES[slug];
  }
  return THEMES.sapling;
}
