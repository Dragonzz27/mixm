export const palette = {
  paper: '#f7f5ef',
  surface: '#fffdf8',
  surfaceMuted: '#efede5',
  ink: '#292923',
  muted: '#706f64',
  faint: '#727064',
  border: '#dedbd0',
  accent: '#a6533b',
  accentHover: '#914631',
  accentPressed: '#783a2a',
  accentSoft: '#f0ded3',
  success: '#5c7054',
  successSoft: '#e7ebdf',
  info: '#527183',
  warning: '#946f32',
  error: '#a6483e'
}

const fontBody =
  '"Inter", "Helvetica Neue", "PingFang SC", "Microsoft YaHei", system-ui, sans-serif'
const fontDisplay =
  '"Iowan Old Style", "Palatino Linotype", "Noto Serif CJK SC", "Songti SC", "SimSun", Georgia, serif'

export const cssVariables = {
  ...Object.fromEntries(
    Object.entries(palette).map(([key, value]) => [
      `--mixm-${key.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)}`,
      value
    ])
  ),
  '--mixm-font-body': fontBody,
  '--mixm-font-display': fontDisplay
}

export const themeOverrides = {
  common: {
    fontFamily: fontBody,
    fontWeightStrong: '600',
    lineHeight: '1.65',
    borderRadius: '5px',
    borderRadiusSmall: '3px',
    primaryColor: palette.accent,
    primaryColorHover: palette.accentHover,
    primaryColorPressed: palette.accentPressed,
    primaryColorSuppl: palette.accentHover,
    successColor: palette.success,
    successColorHover: '#6c805f',
    successColorPressed: '#46583e',
    successColorSuppl: palette.success,
    infoColor: palette.info,
    infoColorHover: '#658596',
    infoColorPressed: '#3d596a',
    infoColorSuppl: palette.info,
    warningColor: palette.warning,
    warningColorHover: '#a98244',
    warningColorPressed: '#755626',
    warningColorSuppl: palette.warning,
    errorColor: palette.error,
    errorColorHover: '#b85d52',
    errorColorPressed: '#86372f',
    errorColorSuppl: palette.error,
    textColorBase: palette.ink,
    textColor1: palette.ink,
    textColor2: palette.ink,
    textColor3: palette.muted,
    placeholderColor: palette.faint,
    borderColor: palette.border,
    dividerColor: palette.border,
    baseColor: palette.surface,
    bodyColor: palette.paper,
    cardColor: palette.surface,
    modalColor: palette.surface,
    popoverColor: palette.surface,
    inputColor: palette.surface,
    inputColorDisabled: palette.surfaceMuted,
    tableColor: palette.surface,
    tableHeaderColor: palette.surfaceMuted,
    tabColor: palette.surfaceMuted,
    actionColor: palette.surfaceMuted,
    hoverColor: palette.surfaceMuted,
    codeColor: palette.surfaceMuted,
    progressRailColor: '#e6e3d8',
    boxShadow1: '0 3px 16px rgb(41 41 35 / 6%)',
    boxShadow2: '0 12px 40px rgb(41 41 35 / 10%)',
    boxShadow3: '0 18px 64px rgb(41 41 35 / 14%)'
  },
  Button: {
    colorPrimary: palette.ink,
    colorHoverPrimary: '#47473e',
    colorPressedPrimary: '#181814',
    colorFocusPrimary: '#47473e',
    textColorPrimary: palette.surface,
    textColorHoverPrimary: palette.surface,
    textColorPressedPrimary: palette.surface,
    textColorFocusPrimary: palette.surface,
    borderPrimary: `1px solid ${palette.ink}`,
    borderHoverPrimary: '1px solid #47473e',
    borderPressedPrimary: '1px solid #181814',
    borderFocusPrimary: '1px solid #47473e',
    fontWeight: '500'
  },
  Card: {
    borderColor: palette.border,
    paddingMedium: '24px',
    titleFontWeight: '500',
    borderRadius: '6px'
  },
  Menu: {
    itemHeight: '42px',
    itemColorActive: palette.accentSoft,
    itemColorActiveHover: palette.accentSoft,
    itemColorActiveCollapsed: palette.accentSoft,
    itemTextColorActive: palette.accentPressed,
    itemTextColorActiveHover: palette.accentPressed,
    itemIconColorActive: palette.accentPressed,
    itemIconColorActiveHover: palette.accentPressed,
    itemTextColor: palette.muted,
    itemIconColor: palette.muted,
    itemColorHover: '#e8e5db',
    borderRadius: '4px'
  },
  DataTable: {
    thColor: palette.surfaceMuted,
    thTextColor: palette.muted,
    thFontWeight: '500',
    tdColor: palette.surface,
    tdColorHover: '#f2efe6',
    borderColor: palette.border,
    borderRadius: '4px'
  },
  Input: { boxShadowFocus: `0 0 0 3px ${palette.accent}1a` },
  Tabs: { tabFontWeightActive: '600', tabBorderRadius: '4px' }
}
