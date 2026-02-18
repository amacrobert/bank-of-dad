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

const saplingLeavesSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='%232D5A3D' opacity='0.07'><path d='M12 8c-2 3-1 6 1 7s5-1 5-4-3-5-6-3z'/><path d='M55 22c-1 2 0 5 2 5s4-2 3-4-3-3-5-1z'/><path d='M30 45c-2 2-1 5 1 6s4 0 4-3-2-5-5-3z'/><path d='M68 58c-1 2 0 4 2 5s3-1 3-3-2-4-5-2z'/><path d='M8 62c-1 2 0 4 2 4s3-1 3-3-3-3-5-1z'/><path d='M45 70c-2 2-1 5 1 5s4-1 4-3-3-4-5-2z'/><path d='M72 15c-1 2 0 4 1 4s3-1 3-3-2-3-4-1z'/><path d='M38 12c-1 1 0 3 1 3s2-1 2-2-1-2-3-1z'/></g></svg>`)}")`;

const piggyBankCoinsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='270' height='270' viewBox='0 0 90 90'><g fill='none' stroke='%238B4560' opacity='0.06' stroke-width='1'><circle cx='15' cy='12' r='6'/><circle cx='60' cy='8' r='4'/><circle cx='40' cy='35' r='7'/><circle cx='78' cy='30' r='5'/><circle cx='10' cy='55' r='5'/><circle cx='55' cy='60' r='6'/><circle cx='30' cy='75' r='4'/><circle cx='75' cy='72' r='5'/></g></svg>`)}")`;

const rainbowStarsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='%235B4BA0' opacity='0.06'><path d='M10 10l2 4 4 1-3 3 1 4-4-2-4 2 1-4-3-3 4-1z'/><path d='M55 8l1 3 3 0-2 2 1 3-3-1-3 1 1-3-2-2 3 0z'/><path d='M35 35l2 4 4 1-3 3 1 4-4-2-4 2 1-4-3-3 4-1z'/><path d='M70 40l1 3 3 0-2 2 1 3-3-1-3 1 1-3-2-2 3 0z'/><path d='M15 60l1 3 3 0-2 2 1 3-3-1-3 1 1-3-2-2 3 0z'/><path d='M60 68l2 4 4 1-3 3 1 4-4-2-4 2 1-4-3-3 4-1z'/></g></svg>`)}")`;

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
  rainbow: {
    label: "Rainbow",
    slug: "rainbow",
    colors: {
      forest: "#5B4BA0",
      forestLight: "#7366B8",
      cream: "#F3F0FD",
      creamDark: "#E8E3F5",
    },
    backgroundSvg: rainbowStarsSvg,
  },
};

export const THEME_SLUGS = Object.keys(THEMES);

export function getTheme(slug: string | null | undefined): ThemeDefinition {
  if (slug && THEMES[slug]) {
    return THEMES[slug];
  }
  return THEMES.sapling;
}
