/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { MultiSelectOption } from "@components/inputs/select";
import type { TFunction } from "i18next";

export const AnnounceTypeOptions: MultiSelectOption[] = [
  {
    label: "New",
    value: "NEW"
  },
  {
    label: "Checked",
    value: "CHECKED"
  },
  {
    label: "Promo",
    value: "PROMO"
  },
  {
    label: "Promo GP",
    value: "PROMO_GP"
  },
  {
    label: "Resurrected",
    value: "RESURRECTED"
  }
];

export const getAnnounceTypeOptions = (t: TFunction): MultiSelectOption[] => [
  { label: t("options:announceType.NEW"), value: "NEW" },
  { label: t("options:announceType.CHECKED"), value: "CHECKED" },
  { label: t("options:announceType.PROMO"), value: "PROMO" },
  { label: t("options:announceType.PROMO_GP"), value: "PROMO_GP" },
  { label: t("options:announceType.RESURRECTED"), value: "RESURRECTED" }
];

export const resolutions = [
  "2160p",
  "1080p",
  "810p",
  "720p",
  "576p",
  "480p"
];

export const RESOLUTION_OPTIONS: MultiSelectOption[] = resolutions.map(r => ({ value: r, label: r, key: r }));

export const codecs = [
  "AV1",
  "AVC",
  "H.264",
  "H.265",
  "HEVC",
  "MPEG-2",
  "VC-1",
  "XviD",
  "x264",
  "x265"
];

export const CODECS_OPTIONS: MultiSelectOption[] = codecs.map(v => ({ value: v, label: v, key: v }));

export const sources = [
  "AHDTV",
  "BD5",
  "BD9",
  "BDRip",
  "BDr",
  "BRRip",
  "BluRay",
  "CAM",
  "DVDR",
  "DVDRip",
  "DVDScr",
  "HDCAM",
  "HDDVD",
  "HDDVDRip",
  "HDTS",
  "HDTV",
  "Mixed",
  "SiteRip",
  "UHD.BluRay",
  "WEB",
  "WEB-DL",
  "WEBRip"
];

export const SOURCES_OPTIONS: MultiSelectOption[] = sources.map(v => ({ value: v, label: v, key: v }));

export const containers = [
  "avi",
  "mp4",
  "mkv"
];

export const CONTAINER_OPTIONS: MultiSelectOption[] = containers.map(v => ({ value: v, label: v, key: v }));

export const hdr = [
  "HDR",
  "HDR10",
  "HDR10+",
  "HLG",
  "DV",
  "DV HDR",
  "DV HDR10",
  "DV HDR10+",
  "DoVi",
  "Dolby Vision"
];

export const HDR_OPTIONS: MultiSelectOption[] = hdr.map(v => ({ value: v, label: v, key: v }));

export const quality_other = [
  "REMUX",
  "HYBRID",
  "REPACK"
];

export const OTHER_OPTIONS = quality_other.map(v => ({ value: v, label: v, key: v }));

export const formatMusic = [
  "MP3",
  "FLAC",
  "Ogg Vorbis",
  "Ogg",
  "AAC",
  "AC3",
  "DTS",
  "DSD"
];

export const FORMATS_OPTIONS: MultiSelectOption[] = formatMusic.map(r => ({ value: r, label: r, key: r }));

export const sourcesMusic = [
  "CD",
  "WEB",
  "DVD",
  "Vinyl",
  "Soundboard",
  "DAT",
  "Cassette",
  "Blu-Ray",
  "SACD"
];

export const SOURCES_MUSIC_OPTIONS: MultiSelectOption[] = sourcesMusic.map(v => ({ value: v, label: v, key: v }));

export const qualityMusic = [
  "192",
  "256",
  "320",
  "APS (VBR)",
  "APX (VBR)",
  "V2 (VBR)",
  "V1 (VBR)",
  "V0 (VBR)",
  "Lossless",
  "24bit Lossless",
  "DSD64",
  "DSD128",
  "DSD256",
  "DSD512"
];

export const QUALITY_MUSIC_OPTIONS: MultiSelectOption[] = qualityMusic.map(v => ({ value: v, label: v, key: v }));

export const releaseTypeMusic = [
  "Album",
  "Anthology",
  "Bootleg",
  "Compilation",
  "Concert Recording",
  "Demo",
  "DJ Mix",
  "EP",
  "Interview",
  "Live album",
  "Mixtape",
  "Remix",
  "Sampler",
  "Single",
  "Soundtrack",
  "Unknown"
];

export const RELEASE_TYPE_MUSIC_OPTIONS: MultiSelectOption[] = releaseTypeMusic.map(v => ({
  value: v,
  label: v,
  key: v
}));

export const formatBooks = [
  "epub",
  "pdf",
  "mobi",
  "azw3",
  "azw",
  "cbz",
  "cbr",
  "pdb",
  "doc",
  "docx",
  "lit",
  "chm",
  "djvu",
  "lrf",
  "fb2",
  "html",
  "htm",
  "txt",
  "rtf",
  "xml",
  "mp3",
  "m4a",
  "m4b",
  "aac",
  "ogg",
  "wma",
  "flac",
  "wav"
];

export const FORMATS_BOOKS_OPTIONS: MultiSelectOption[] = formatBooks.map(r => ({ value: r, label: r, key: r }));

export const originOptions = [
  "P2P",
  "Internal",
  "SCENE",
  "O-SCENE"
];

export const ORIGIN_OPTIONS = originOptions.map(v => ({ value: v, label: v, key: v }));

export const languageOptions = [
  "BALTIC",
  "BRAZiLiAN",
  "BULGARiAN",
  "CHiNESE",
  "CHS",
  "CHT",
  "CZECH",
  "DANiSH",
  "DUBBED",
  "DKSUBS",
  "DUTCH",
  "ENGLiSH",
  "ESTONiAN",
  "FLEMiSH",
  "FiNNiSH",
  "FRENCH",
  "GERMAN",
  "GREEK",
  "HAiTiAN",
  "HARDSUB",
  "Hardcoded",
  "HEBREW",
  "HebSub",
  "HiNDi",
  "HUNGARiAN",
  "iCELANDiC",
  "iTALiAN",
  "JAPANESE",
  "KOREAN",
  "LATiN",
  "MANDARiN",
  "MULTi",
  "MULTILANG",
  "MULTiSUB",
  "MULTiSUBS",
  "NORDiC",
  "NORWEGiAN",
  "POLiSH",
  "PORTUGUESE",
  "ROMANiAN",
  "RUSSiAN",
  "SLOVAK",
  "SPANiSH",
  "SUBBED",
  "SUBFORCED",
  "SUBPACK",
  "SWEDiSH",
  "SYNCED",
  "TURKiSH",
  "UKRAiNiAN",
  "UNSUBBED"
];

export const LANGUAGE_OPTIONS = languageOptions.map(v => ({ value: v, label: v, key: v }));

export interface RadioFieldsetOption {
  label: string;
  description: string;
  value: ActionType;
  type?: string;
}

export const ActionTypeOptions: RadioFieldsetOption[] = [
  { label: "Test", description: "A simple action to test a filter.", value: "TEST" },
  { label: "Watch dir", description: "Add filtered torrents to a watch directory", value: "WATCH_FOLDER" },
  { label: "Webhook", description: "Run webhook", value: "WEBHOOK" },
  { label: "Exec", description: "Run a custom command after a filter match", value: "EXEC" }
];

export const getActionTypeOptions = (t: TFunction): RadioFieldsetOption[] => [
  { label: t("options:actionType.TEST.label"), description: t("options:actionType.TEST.description"), value: "TEST" },
  { label: t("options:actionType.WATCH_FOLDER.label"), description: t("options:actionType.WATCH_FOLDER.description"), value: "WATCH_FOLDER" },
  { label: t("options:actionType.WEBHOOK.label"), description: t("options:actionType.WEBHOOK.description"), value: "WEBHOOK" },
  { label: t("options:actionType.EXEC.label"), description: t("options:actionType.EXEC.description"), value: "EXEC" }
];

export const ActionTypeNameMap: Record<ActionType, string> = {
  "TEST": "Test",
  "WATCH_FOLDER": "Watch folder",
  "WEBHOOK": "Webhook",
  "EXEC": "Exec"
} as const;

export const getActionTypeNameMap = (t: TFunction): Record<ActionType, string> => ({
  "TEST": t("options:actionType.TEST.label"),
  "WATCH_FOLDER": t("options:actionType.WATCH_FOLDER.label"),
  "WEBHOOK": t("options:actionType.WEBHOOK.label"),
  "EXEC": t("options:actionType.EXEC.label")
});

export interface OptionBasic {
  label: string;
  value: string;
}

export interface OptionBasicTyped<T> {
  label: string;
  value: T;
}

export const PushStatusOptions: OptionBasic[] = [
  {
    label: "Rejected",
    value: "PUSH_REJECTED"
  },
  {
    label: "Approved",
    value: "PUSH_APPROVED"
  },
  {
    label: "Pending",
    value: "PENDING"
  },
  {
    label: "Error",
    value: "PUSH_ERROR"
  }
];

export const getPushStatusOptions = (t: TFunction): OptionBasic[] => [
  { label: t("options:pushStatus.PUSH_REJECTED"), value: "PUSH_REJECTED" },
  { label: t("options:pushStatus.PUSH_APPROVED"), value: "PUSH_APPROVED" },
  { label: t("options:pushStatus.PENDING"), value: "PENDING" },
  { label: t("options:pushStatus.PUSH_ERROR"), value: "PUSH_ERROR" }
];

export const ListTypeOptions: OptionBasicTyped<ListType>[] = [
  {
    label: "Trakt",
    value: "TRAKT"
  },
  {
    label: "Plaintext",
    value: "PLAINTEXT"
  },
  {
    label: "AniList",
    value: "ANILIST"
  }
];

export const ListTypeNameMap: Record<ListType, string> = {
  "TRAKT": "Trakt",
  "PLAINTEXT": "Plaintext",
  "ANILIST": "AniList",
};

export const getNotificationTypeOptions = (t: TFunction): OptionBasicTyped<NotificationType>[] => [
  { label: t("options:notificationType.WEBHOOK"), value: "WEBHOOK" }
];

export const IrcAuthMechanismTypeOptions: OptionBasicTyped<IrcAuthMechanism>[] = [
  {
    label: "None",
    value: "NONE"
  },
  {
    label: "SASL (plain)",
    value: "SASL_PLAIN"
  },
  {
    label: "NickServ",
    value: "NICKSERV"
  }
];

export const downloadsPerUnitOptions: OptionBasic[] = [
  {
    label: "Select",
    value: ""
  },
  {
    label: "MINUTE",
    value: "MINUTE"
  },
  {
    label: "HOUR",
    value: "HOUR"
  },
  {
    label: "DAY",
    value: "DAY"
  },
  {
    label: "WEEK",
    value: "WEEK"
  },
  {
    label: "MONTH",
    value: "MONTH"
  },
  {
    label: "EVER",
    value: "EVER"
  }
];

export const getWindowTypeOptions = (t: TFunction): OptionBasic[] => [
  { label: t("options:maxDownloadsWindowType.FIXED"), value: "FIXED" },
  { label: t("options:maxDownloadsWindowType.ROLLING"), value: "ROLLING" }
];

const logLevel = ["DEBUG", "INFO", "WARN", "ERROR", "TRACE"] as const;

export const LogLevelOptions = logLevel.map(v => ({ value: v, label: v, key: v }));

export interface SelectOption {
  label: string;
  description: string;
  value: NotificationEvent;
}

export interface SelectGenericOption<T> {
  label: string;
  description: string;
  value: T;
}

export const EventOptions: SelectOption[] = [
  {
    label: "Push Rejected",
    value: "PUSH_REJECTED",
    description: "On push rejected for the arrs or download client"
  },
  {
    label: "Push Approved",
    value: "PUSH_APPROVED",
    description: "On push approved for the arrs or download client"
  },
  {
    label: "Push Error",
    value: "PUSH_ERROR",
    description: "On push error for the arrs or download client"
  },
  {
    label: "IRC Disconnected",
    value: "IRC_DISCONNECTED",
    description: "Unexpectedly disconnected from irc network"
  },
  {
    label: "IRC Reconnected",
    value: "IRC_RECONNECTED",
    description: "Reconnected to irc network after error"
  },
  {
    label: "New update",
    value: "APP_UPDATE_AVAILABLE",
    description: "Get notified on updates"
  },
  {
    label: "New Release",
    value: "RELEASE_NEW",
    description: "On new release from indexer (before filtering)"
  }
];

export const getEventOptions = (t: TFunction): SelectOption[] => [
  {
    label: t("options:event.PUSH_REJECTED.label"),
    value: "PUSH_REJECTED",
    description: t("options:event.PUSH_REJECTED.description")
  },
  {
    label: t("options:event.PUSH_APPROVED.label"),
    value: "PUSH_APPROVED",
    description: t("options:event.PUSH_APPROVED.description")
  },
  {
    label: t("options:event.PUSH_ERROR.label"),
    value: "PUSH_ERROR",
    description: t("options:event.PUSH_ERROR.description")
  },
  {
    label: t("options:event.IRC_DISCONNECTED.label"),
    value: "IRC_DISCONNECTED",
    description: t("options:event.IRC_DISCONNECTED.description")
  },
  {
    label: t("options:event.IRC_RECONNECTED.label"),
    value: "IRC_RECONNECTED",
    description: t("options:event.IRC_RECONNECTED.description")
  },
  {
    label: t("options:event.IRC_UNHEALTHY.label"),
    value: "IRC_UNHEALTHY",
    description: t("options:event.IRC_UNHEALTHY.description")
  },
  {
    label: t("options:event.IRC_HEALTHY.label"),
    value: "IRC_HEALTHY",
    description: t("options:event.IRC_HEALTHY.description")
  },
  {
    label: t("options:event.LIST_REFRESH_SUCCESS.label"),
    value: "LIST_REFRESH_SUCCESS",
    description: t("options:event.LIST_REFRESH_SUCCESS.description")
  },
  {
    label: t("options:event.LIST_REFRESH_ERROR.label"),
    value: "LIST_REFRESH_ERROR",
    description: t("options:event.LIST_REFRESH_ERROR.description")
  },
  {
    label: t("options:event.FEED_REFRESH_SUCCESS.label"),
    value: "FEED_REFRESH_SUCCESS",
    description: t("options:event.FEED_REFRESH_SUCCESS.description")
  },
  {
    label: t("options:event.FEED_REFRESH_ERROR.label"),
    value: "FEED_REFRESH_ERROR",
    description: t("options:event.FEED_REFRESH_ERROR.description")
  },
  {
    label: t("options:event.APP_UPDATE_AVAILABLE.label"),
    value: "APP_UPDATE_AVAILABLE",
    description: t("options:event.APP_UPDATE_AVAILABLE.description")
  },
  {
    label: t("options:event.RELEASE_NEW.label"),
    value: "RELEASE_NEW",
    description: t("options:event.RELEASE_NEW.description")
  }
];

export const FeedDownloadTypeOptions: OptionBasicTyped<FeedDownloadType>[] = [
  {
    label: "Magnet",
    value: "MAGNET"
  },
  {
    label: "Torrent",
    value: "TORRENT"
  }
];

export const tagsMatchLogicOptions: OptionBasic[] = [
  {
    label: "any",
    value: "ANY"
  },
  {
    label: "all",
    value: "ALL"
  }
];

export const ExternalFilterTypeOptions: RadioFieldsetOption[] = [
  { label: "Exec", description: "Run a custom command", value: "EXEC" },
  { label: "Webhook", description: "Run webhook", value: "WEBHOOK" }
];

export const ExternalFilterTypeNameMap = {
  "EXEC": "Exec",
  "WEBHOOK": "Webhook"
};

export const ExternalFilterOnErrorValues: ExternalFilterOnError[] = ["REJECT", "CONTINUE"];

export const ExternalFilterOnErrorOptions: SelectGenericOption<ExternalFilterOnError>[] = [
  { label: "Reject", description: "Reject on error, default", value: "REJECT" },
  { label: "Continue to next", description: "Continue on error", value: "CONTINUE" }
];

export const ExternalFilterWebhookMethodOptions: OptionBasicTyped<WebhookMethod>[] = [
  { label: "GET", value: "GET" },
  { label: "POST", value: "POST" },
  { label: "PUT", value: "PUT" },
  { label: "PATCH", value: "PATCH" },
  { label: "DELETE", value: "DELETE" }
];

export const ProxyTypeOptions: OptionBasicTyped<ProxyType>[] = [
  {
    label: "SOCKS5",
    value: "SOCKS5"
  },
  {
    label: "HTTP",
    value: "HTTP"
  }
];

export const ListsTraktOptions: OptionBasic[] = [
  {
    label: "Anticipated TV",
    value: "https://api.autobrr.com/lists/trakt/anticipated-tv"
  },
  {
    label: "Popular TV",
    value: "https://api.autobrr.com/lists/trakt/popular-tv"
  },
  {
    label: "Upcoming Movies",
    value: "https://api.autobrr.com/lists/trakt/upcoming-movies"
  },
  {
    label: "Upcoming BluRay",
    value: "https://api.autobrr.com/lists/trakt/upcoming-bluray"
  },
  {
    label: "Popular TV",
    value: "https://api.autobrr.com/lists/trakt/popular-tv"
  },
  {
    label: "Steven Lu",
    value: "https://api.autobrr.com/lists/stevenlu"
  }
];

export const ListsAniListOptions: OptionBasic[] = [
  {
    label: "Current anime season",
    value: "https://api.autobrr.com/lists/anilist/seasonal"
  },
  {
    label: "Trending animes",
    value: "https://api.autobrr.com/lists/anilist/trending"
  },
  {
    label: "Next anime season",
    value: "https://api.autobrr.com/lists/anilist/upcoming"
  }
];
