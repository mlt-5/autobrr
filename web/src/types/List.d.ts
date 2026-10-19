/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

interface List {
  id: number;
  name: string;
  enabled: boolean;
  type: ListType;
  url: string;
  headers: string[];
  api_key: string;
  filters: ListFilter[];
  match_release: boolean;
  include_year: boolean;
  skip_clean_sanitize: boolean;
  last_refresh_status: string;
}

interface ListFilter {
  id: number;
  name: string;
}

interface ListCreate {
  name: string;
  enabled: boolean;
  type: ListType;
  url: string;
  headers: string[];
  api_key: string;
  filters: number[];
  match_release: boolean;
  include_year: boolean;
  skip_clean_sanitize: boolean;
}

type ListType =
  | "MDBLIST"
  | "TRAKT"
  | "METACRITIC"
  | "STEAM"
  | "PLAINTEXT"
  | "ANILIST";
