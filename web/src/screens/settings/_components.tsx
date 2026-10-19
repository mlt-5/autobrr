/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { classNames } from "@utils";
import { SVGProps } from "react";

type SectionProps = {
  title: string;
  titleElement?: React.ReactNode;
  description: string | React.ReactNode;
  rightSide?: React.ReactNode;
  children?: React.ReactNode;
  noLeftPadding?: boolean;
};

export const Section = ({
  title,
  titleElement,
  description,
  rightSide,
  children,
  noLeftPadding = false,
}: SectionProps) => (
  <div
    className={classNames(
      "pb-6 px-4 lg:col-span-9",
      noLeftPadding ? 'pl-0' : '',
    )}
  >
    <div
      className={classNames(
        "mt-6 mb-4",
        rightSide
          ? "flex justify-between items-start flex-wrap sm:flex-nowrap gap-2"
          : ""
      )}
    >
      <div className="sm:px-2">
        {titleElement ?? <h2 className="text-lg leading-4 font-bold text-gray-900 dark:text-white">{title}</h2>}
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">{description}</p>
      </div>
      <div className="shrink-0">
        {rightSide ?? null}
      </div>
    </div>
    {children}
  </div>
);

interface RowItemProps {
  label: string;
  value?: string | React.ReactNode;
  title?: string;
  emptyText?: string;
  rightSide?: React.ReactNode;
  className?: string;
}

export const RowItem = ({
  label,
  value,
  title,
  emptyText,
  rightSide,
  className = "sm:col-span-3"
}: RowItemProps) => (
  <div className="p-4 sm:px-6 sm:grid sm:grid-cols-4 sm:gap-4">
    <div className="font-medium text-gray-900 dark:text-white text-sm self-center" title={title}>
      {label}
    </div>
    <div
      className={classNames(
        className,
        "mt-1 text-gray-900 dark:text-gray-300 text-sm break-all sm:mt-0"
      )}
    >
      {value
        ? (
          <>
            {typeof (value) === "string" ? (
              <span className="px-1.5 py-1 bg-gray-200 dark:bg-gray-700 rounded-sm shadow-sm text-ellipsis leading-7">
                {value}
              </span>
            ) : value}
            {rightSide ?? null}
          </>
        )
        : (emptyText ?? null)
      }
    </div>
  </div>
);

const commonSVGProps: SVGProps<SVGSVGElement> = {
  clipRule: "evenodd", fill: "currentColor", fillRule: "evenodd", xmlns: "http://www.w3.org/2000/svg",
  className: "mr-2 h-5"
};

export const WebhookIcon = () => (
  <svg
    {...commonSVGProps}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M18 16.98h-5.99c-1.1 0-1.95.94-2.48 1.9A4 4 0 0 1 2 17c.01-.7.2-1.4.57-2" />
    <path d="m6 17 3.13-5.78c.53-.97.1-2.18-.5-3.1a4 4 0 1 1 6.89-4.06" />
    <path d="m12 6 3.13 5.73C15.66 12.7 16.9 13 18 13a4 4 0 0 1 0 8" />
  </svg>
)


