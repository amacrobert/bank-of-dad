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

const seafoamShellsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#1A7A6D' opacity='0.07'><path d='M12 14c0-4 3-7 6-7s5 3 5 7c0 3-2 5-5 5s-6-2-6-5z' /><path d='M10 14h12M12 10c2 3 2 6 0 9M18 10c-2 3-2 6 0 9'/><path d='M55 10c0-3 2-5 4-5s4 2 4 5c0 2-2 4-4 4s-4-2-4-4z'/><path d='M30 42c0-5 4-8 7-8s6 3 6 8c0 4-3 6-6 6s-7-2-7-6z' /><path d='M28 42h14M31 37c2 4 2 7 0 11M39 37c-2 4-2 7 0 11'/><path d='M68 32c0-3 2-5 4-5s4 2 4 5-2 4-4 4-4-2-4-4z'/><path d='M10 68c0-3 2-5 4-5s4 2 4 5-2 4-4 4-4-2-4-4z'/><path d='M55 65c0-4 3-7 6-7s5 3 5 7c0 3-2 5-5 5s-6-2-6-5z'/><path d='M53 65h14M56 61c2 3 2 6 0 9M62 61c-2 3-2 6 0 9'/></g></svg>`)}")`;

const dinostompFootprintsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#5C6B2F' opacity='0.07'><ellipse cx='14' cy='16' rx='4' ry='6'/><circle cx='9' cy='8' r='1.8'/><circle cx='14' cy='7' r='1.8'/><circle cx='19' cy='8' r='1.8'/><ellipse cx='58' cy='12' rx='3' ry='5' transform='rotate(-15 58 12)'/><circle cx='54' cy='5' r='1.5'/><circle cx='58' cy='4' r='1.5'/><circle cx='62' cy='5' r='1.5'/><ellipse cx='35' cy='42' rx='4.5' ry='7'/><circle cx='30' cy='33' r='2'/><circle cx='35' cy='32' r='2'/><circle cx='40' cy='33' r='2'/><ellipse cx='72' cy='55' rx='3' ry='5'/><circle cx='68' cy='48' r='1.5'/><circle cx='72' cy='47' r='1.5'/><circle cx='76' cy='48' r='1.5'/><ellipse cx='15' cy='72' rx='3.5' ry='5.5' transform='rotate(10 15 72)'/><circle cx='11' cy='64' r='1.7'/><circle cx='15' cy='63' r='1.7'/><circle cx='19' cy='64' r='1.7'/><ellipse cx='55' cy='75' rx='3' ry='4.5'/><circle cx='52' cy='69' r='1.3'/><circle cx='55' cy='68' r='1.3'/><circle cx='58' cy='69' r='1.3'/></g></svg>`)}")`;

const campfireTentsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#B5541A' opacity='0.07'><path d='M10 22L16 8l6 14H10z'/><path d='M50 18L55 6l5 12H50z'/><path d='M60 50l-5 10h10z'/><path d='M62 47v-7l3 4z'/><path d='M58 47v-5l-2 3z'/><path d='M30 45L37 28l7 17H30z'/><path d='M5 60l-3 10h6z'/><path d='M7 57v-7l3 4z'/><path d='M3 57v-5l-2 3z'/><path d='M70 72l-5 8h10z'/><path d='M72 68v-6l3 3.5z'/><path d='M68 68v-4l-2 2.5z'/><path d='M20 75l-3 5h6z'/><path d='M22 72v-5l2 3z'/><path d='M18 72v-4l-2 2z'/></g></svg>`)}")`;

const ninjaShurikensSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#3A3A3A' opacity='0.07'><path d='M14 14l-5-2 2-5 5 2-2 5zM14 14l5 2-2 5-5-2 2-5z'/><path d='M58 10l-4-1.5 1.5-4 4 1.5-1.5 4zM58 10l4 1.5-1.5 4-4-1.5 1.5-4z'/><path d='M35 38l-6-2.5 2.5-6 6 2.5-2.5 6zM35 38l6 2.5-2.5 6-6-2.5 2.5-6z'/><path d='M72 45l-4-1.5 1.5-4 4 1.5-1.5 4zM72 45l4 1.5-1.5 4-4-1.5 1.5-4z'/><path d='M12 62l-5-2 2-5 5 2-2 5zM12 62l5 2-2 5-5-2 2-5z'/><path d='M58 70l-4-1.5 1.5-4 4 1.5-1.5 4zM58 70l4 1.5-1.5 4-4-1.5 1.5-4z'/></g></svg>`)}")`;

const arcticSnowflakesSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g stroke='#3B7CA5' fill='none' stroke-width='1' opacity='0.07'><g transform='translate(14,14)'><line x1='0' y1='-7' x2='0' y2='7'/><line x1='-6' y1='-3.5' x2='6' y2='3.5'/><line x1='-6' y1='3.5' x2='6' y2='-3.5'/><line x1='-2' y1='-4.5' x2='2' y2='-6.5'/><line x1='-2' y1='4.5' x2='2' y2='6.5'/></g><g transform='translate(55,12)'><line x1='0' y1='-5' x2='0' y2='5'/><line x1='-4.3' y1='-2.5' x2='4.3' y2='2.5'/><line x1='-4.3' y1='2.5' x2='4.3' y2='-2.5'/></g><g transform='translate(35,40)'><line x1='0' y1='-8' x2='0' y2='8'/><line x1='-7' y1='-4' x2='7' y2='4'/><line x1='-7' y1='4' x2='7' y2='-4'/><line x1='-2.5' y1='-5.5' x2='2.5' y2='-7.5'/><line x1='-2.5' y1='5.5' x2='2.5' y2='7.5'/></g><g transform='translate(70,52)'><line x1='0' y1='-5' x2='0' y2='5'/><line x1='-4.3' y1='-2.5' x2='4.3' y2='2.5'/><line x1='-4.3' y1='2.5' x2='4.3' y2='-2.5'/></g><g transform='translate(12,65)'><line x1='0' y1='-6' x2='0' y2='6'/><line x1='-5.2' y1='-3' x2='5.2' y2='3'/><line x1='-5.2' y1='3' x2='5.2' y2='-3'/></g><g transform='translate(55,72)'><line x1='0' y1='-7' x2='0' y2='7'/><line x1='-6' y1='-3.5' x2='6' y2='3.5'/><line x1='-6' y1='3.5' x2='6' y2='-3.5'/></g></g></svg>`)}")`;

const treasuremapXCompassSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#A07430' opacity='0.07'><path d='M10 10l4 4M14 10l-4 4' stroke='#A07430' stroke-width='2.5' stroke-linecap='round'/><path d='M56 8l3 3M59 8l-3 3' stroke='#A07430' stroke-width='2' stroke-linecap='round'/><circle cx='36' cy='38' r='6' fill='none' stroke='#A07430' stroke-width='1.5'/><path d='M36 31v-2M36 47v-2M29 38h-2M45 38h-2' stroke='#A07430' stroke-width='1.2'/><path d='M36 34l-2 4 2 4 2-4z'/><path d='M68 55l3 3M71 55l-3 3' stroke='#A07430' stroke-width='2' stroke-linecap='round'/><path d='M12 68l4 4M16 68l-4 4' stroke='#A07430' stroke-width='2.5' stroke-linecap='round'/><circle cx='58' cy='72' r='4.5' fill='none' stroke='#A07430' stroke-width='1.2'/><path d='M58 67v-1.5M58 78.5v-1.5M53 72h-1.5M64.5 72h-1.5' stroke='#A07430' stroke-width='1'/><path d='M58 70l-1.5 2 1.5 2 1.5-2z'/></g></svg>`)}")`;

const unicornRainbowsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g opacity='0.07'><path d='M5 20a12 12 0 0 1 24 0' fill='none' stroke='#9B6EC5' stroke-width='2'/><path d='M8 20a9 9 0 0 1 18 0' fill='none' stroke='#D48CBE' stroke-width='2'/><path d='M11 20a6 6 0 0 1 12 0' fill='none' stroke='#E8A87C' stroke-width='2'/><path d='M50 15a8 8 0 0 1 16 0' fill='none' stroke='#9B6EC5' stroke-width='1.5'/><path d='M52 15a6 6 0 0 1 12 0' fill='none' stroke='#D48CBE' stroke-width='1.5'/><g fill='#9B6EC5'><path d='M70 38c2-1 3 0 3 2s-2 3-3 3-3-1-3-3c0-1.5 1.5-2 3-2z'/><path d='M68 44c-1-.5-2 .5-2 1s1 1 2 1 2 0 2-1-.5-1.5-2-1z'/><path d='M14 52c2-1 3 0 3 2s-2 3-3 3-3-1-3-3c0-1.5 1.5-2 3-2z'/><path d='M12 58c-1-.5-2 .5-2 1s1 1 2 1 2 0 2-1-.5-1.5-2-1z'/></g><path d='M30 62a14 14 0 0 1 28 0' fill='none' stroke='#9B6EC5' stroke-width='2'/><path d='M34 62a10 10 0 0 1 20 0' fill='none' stroke='#D48CBE' stroke-width='2'/><path d='M38 62a6 6 0 0 1 12 0' fill='none' stroke='#E8A87C' stroke-width='2'/></g></svg>`)}")`;

const flowerpowerDaisiesSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#C75B5B' opacity='0.07'><g transform='translate(14,14)'><ellipse cx='0' cy='-5' rx='2.5' ry='4'/><ellipse cx='4.8' cy='-1.5' rx='2.5' ry='4' transform='rotate(72)'/><ellipse cx='3' cy='4' rx='2.5' ry='4' transform='rotate(144)'/><ellipse cx='-3' cy='4' rx='2.5' ry='4' transform='rotate(216)'/><ellipse cx='-4.8' cy='-1.5' rx='2.5' ry='4' transform='rotate(288)'/><circle cx='0' cy='0' r='2.5'/></g><g transform='translate(58,10)'><ellipse cx='0' cy='-4' rx='2' ry='3'/><ellipse cx='3.8' cy='-1.2' rx='2' ry='3' transform='rotate(72)'/><ellipse cx='2.4' cy='3.2' rx='2' ry='3' transform='rotate(144)'/><ellipse cx='-2.4' cy='3.2' rx='2' ry='3' transform='rotate(216)'/><ellipse cx='-3.8' cy='-1.2' rx='2' ry='3' transform='rotate(288)'/><circle cx='0' cy='0' r='2'/></g><g transform='translate(35,42)'><circle cx='0' cy='0' r='5'/><circle cx='0' cy='-8' r='2.2'/><circle cx='7.6' cy='-2.5' r='2.2'/><circle cx='4.7' cy='6.5' r='2.2'/><circle cx='-4.7' cy='6.5' r='2.2'/><circle cx='-7.6' cy='-2.5' r='2.2'/></g><g transform='translate(72,50)'><ellipse cx='0' cy='-3.5' rx='1.8' ry='2.5'/><ellipse cx='3.3' cy='-1' rx='1.8' ry='2.5' transform='rotate(72)'/><ellipse cx='2' cy='2.8' rx='1.8' ry='2.5' transform='rotate(144)'/><ellipse cx='-2' cy='2.8' rx='1.8' ry='2.5' transform='rotate(216)'/><ellipse cx='-3.3' cy='-1' rx='1.8' ry='2.5' transform='rotate(288)'/><circle cx='0' cy='0' r='1.8'/></g><g transform='translate(12,68)'><ellipse cx='0' cy='-4' rx='2' ry='3'/><ellipse cx='3.8' cy='-1.2' rx='2' ry='3' transform='rotate(72)'/><ellipse cx='2.4' cy='3.2' rx='2' ry='3' transform='rotate(144)'/><ellipse cx='-2.4' cy='3.2' rx='2' ry='3' transform='rotate(216)'/><ellipse cx='-3.8' cy='-1.2' rx='2' ry='3' transform='rotate(288)'/><circle cx='0' cy='0' r='2'/></g><g transform='translate(60,72)'><circle cx='0' cy='0' r='4'/><circle cx='0' cy='-6.5' r='1.8'/><circle cx='6.2' cy='-2' r='1.8'/><circle cx='3.8' cy='5.3' r='1.8'/><circle cx='-3.8' cy='5.3' r='1.8'/><circle cx='-6.2' cy='-2' r='1.8'/></g></g></svg>`)}")`;

const kittenPawsSvg = `url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' width='240' height='240' viewBox='0 0 80 80'><g fill='#B8705A' opacity='0.07'><g transform='translate(14,14)'><ellipse cx='0' cy='2' rx='3.5' ry='3'/><circle cx='-3.5' cy='-2.5' r='1.8'/><circle cx='-1' cy='-4' r='1.8'/><circle cx='2' cy='-4' r='1.8'/><circle cx='4.5' cy='-2.5' r='1.8'/></g><g transform='translate(55,10)'><ellipse cx='0' cy='1.5' rx='2.8' ry='2.5'/><circle cx='-3' cy='-2' r='1.5'/><circle cx='-0.8' cy='-3.2' r='1.5'/><circle cx='1.5' cy='-3.2' r='1.5'/><circle cx='3.8' cy='-2' r='1.5'/></g><g transform='translate(35,38)'><ellipse cx='0' cy='2.5' rx='4' ry='3.5'/><circle cx='-4' cy='-3' r='2'/><circle cx='-1.2' cy='-4.5' r='2'/><circle cx='2' cy='-4.5' r='2'/><circle cx='5' cy='-3' r='2'/></g><circle cx='70' cy='30' r='3.5'/><path d='M66 30a5 5 0 0 0 8 0' fill='none' stroke='#B8705A' stroke-width='1'/><circle cx='72' cy='35' r='2.5'/><circle cx='67' cy='34' r='2'/><g transform='translate(12,62)'><ellipse cx='0' cy='2' rx='3.2' ry='2.8'/><circle cx='-3.2' cy='-2.2' r='1.7'/><circle cx='-1' cy='-3.5' r='1.7'/><circle cx='1.5' cy='-3.5' r='1.7'/><circle cx='4' cy='-2.2' r='1.7'/></g><g transform='translate(58,68)'><ellipse cx='0' cy='2' rx='3.5' ry='3'/><circle cx='-3.5' cy='-2.5' r='1.8'/><circle cx='-1' cy='-4' r='1.8'/><circle cx='2' cy='-4' r='1.8'/><circle cx='4.5' cy='-2.5' r='1.8'/></g></g></svg>`)}")`;

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
  seafoam: {
    label: "Seafoam",
    slug: "seafoam",
    colors: {
      forest: "#1A7A6D",
      forestLight: "#28A08F",
      cream: "#EFFAF8",
      creamDark: "#D8F0EC",
    },
    backgroundSvg: seafoamShellsSvg,
  },
  dinostomp: {
    label: "Dino Stomp",
    slug: "dinostomp",
    colors: {
      forest: "#5C6B2F",
      forestLight: "#7A8C42",
      cream: "#F6F5EC",
      creamDark: "#EAE8D5",
    },
    backgroundSvg: dinostompFootprintsSvg,
  },
  campfire: {
    label: "Campfire",
    slug: "campfire",
    colors: {
      forest: "#B5541A",
      forestLight: "#D4702E",
      cream: "#FEF5EE",
      creamDark: "#FCE8D5",
    },
    backgroundSvg: campfireTentsSvg,
  },
  ninja: {
    label: "Ninja",
    slug: "ninja",
    colors: {
      forest: "#3A3A3A",
      forestLight: "#555555",
      cream: "#F2F2F2",
      creamDark: "#E0E0E0",
    },
    backgroundSvg: ninjaShurikensSvg,
  },
  arctic: {
    label: "Arctic",
    slug: "arctic",
    colors: {
      forest: "#3B7CA5",
      forestLight: "#5098C0",
      cream: "#EEF6FB",
      creamDark: "#D8EAF4",
    },
    backgroundSvg: arcticSnowflakesSvg,
  },
  treasuremap: {
    label: "Treasure Map",
    slug: "treasuremap",
    colors: {
      forest: "#A07430",
      forestLight: "#C09040",
      cream: "#FDF8EE",
      creamDark: "#F5ECD5",
    },
    backgroundSvg: treasuremapXCompassSvg,
  },
  unicorn: {
    label: "Unicorn",
    slug: "unicorn",
    colors: {
      forest: "#9B6EC5",
      forestLight: "#B388D9",
      cream: "#F8F2FD",
      creamDark: "#EDE3F7",
    },
    backgroundSvg: unicornRainbowsSvg,
  },
  flowerpower: {
    label: "Flower Power",
    slug: "flowerpower",
    colors: {
      forest: "#C75B5B",
      forestLight: "#DE7575",
      cream: "#FDF2F2",
      creamDark: "#F7DEDE",
    },
    backgroundSvg: flowerpowerDaisiesSvg,
  },
  kitten: {
    label: "Kitten",
    slug: "kitten",
    colors: {
      forest: "#B8705A",
      forestLight: "#D08A72",
      cream: "#FDF5F0",
      creamDark: "#F5E5DA",
    },
    backgroundSvg: kittenPawsSvg,
  },
};

export const THEME_SLUGS = Object.keys(THEMES);

export function getTheme(slug: string | null | undefined): ThemeDefinition {
  if (slug && THEMES[slug]) {
    return THEMES[slug];
  }
  return THEMES.sapling;
}
