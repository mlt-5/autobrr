/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { useMutation } from "@tanstack/react-query";
import { useSelector } from "@tanstack/react-form";
import { UserIcon } from "@heroicons/react/24/solid";
import { useTranslation } from "react-i18next";

import { APIClient } from "@api/APIClient";
import { fieldErrors, useAppForm } from "@hooks/form";
import { Section } from "./_components";
import { PasswordField, TextField } from "@components/inputs";
import toast from "@components/hot-toast";
import Toast from "@components/notifications/Toast";
import { AuthContext } from "@utils/Context";

const AccountSettings = () => {
  const { t } = useTranslation("settings");

  return (
    <Section
      title={t("account.title")}
      description={t("account.description")}
    >
      <div className="py-0.5">
        <Credentials />
      </div>
    </Section>
  );
};

interface InputValues {
  username: string;
  newUsername: string;
  oldPassword: string;
  newPassword: string;
  confirmPassword: string;
}

function Credentials() {
  const { t } = useTranslation("settings");
  const username = AuthContext.useSelector((s) => s.username);

  const validate = (values: InputValues) => {
    const errors: Record<string, string> = {};

    if (!values.username)
      errors.username = t("account.required");

    if (values.newPassword !== values.confirmPassword)
      errors.confirmPassword = t("account.passwordsDoNotMatch");

    return errors;
  };

  const logoutMutation = useMutation({
    mutationFn: APIClient.auth.logout,
    onSuccess: () => {
      AuthContext.reset();

      toast.custom((tst) => (
        <Toast type="success" body={t("account.updateSuccess")} t={tst} />
      ));
    }
  });

  const updateUserMutation = useMutation({
    mutationFn: (data: UserUpdate) => APIClient.auth.updateUser(data),
    onError: () => {
      toast.custom((tst) => (
        <Toast type="error" body={t("account.updateError")} t={tst} />
      ));
    },
    onSuccess: () => {
      logoutMutation.mutate();
    }
  });

  const form = useAppForm({
    defaultValues: {
      username: username,
      newUsername: "",
      oldPassword: "",
      newPassword: "",
      confirmPassword: ""
    },
    validators: {
      onChange: ({ value }) => fieldErrors(validate(value))
    },
    onSubmit: ({ value }) => {
      updateUserMutation.mutate({
        username_current: value.username,
        username_new: value.newUsername,
        password_current: value.oldPassword,
        password_new: value.newPassword,
      });
    }
  });

  const newPassword = useSelector(form.store, (state) => state.values.newPassword);

  const separatorClass = "mb-6";

  return (
    <Section
      title={t("account.changeCredentialsTitle")}
      description={t("account.changeCredentialsDescription")}
      noLeftPadding
    >
      <div className="px-2 pb-0 sm:pb-6 bg-white dark:bg-gray-800">
        <form.AppForm>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              e.stopPropagation();
              form.handleSubmit();
            }}
          >
            <div className="flex flex-col sm:grid sm:grid-cols-2 gap-x-10 pt-2">
              <div className={separatorClass}>
                <TextField name="username" label={t("account.currentUsername")} autoComplete="username" disabled />
              </div>
              <div className={separatorClass}>
                <TextField name="newUsername" label={t("account.newUsername")} tooltip={
                  <div>
                    <p>{t("account.optional")}</p>
                  </div>
                } />
              </div>

              <hr className="col-span-2 mb-6 border-t border-gray-300 dark:border-gray-750" />

              <div className={separatorClass}>
                <PasswordField name="oldPassword" placeholder={t("account.required")} label={t("account.currentPassword")} autoComplete="current-password" required tooltip={
                  <div>
                    <p>{t("account.requiredIfUpdatingCredentials")}</p>
                  </div>
                } />
              </div>
              <div>
                <div className={separatorClass}>
                  <PasswordField name="newPassword" label={t("account.newPassword")} autoComplete="new-password" tooltip={
                    <div>
                      <p>{t("account.optional")}</p>
                    </div>
                  } />
                </div>
                {newPassword && (
                  <div className={separatorClass}>
                    <PasswordField name="confirmPassword" label={t("account.confirmNewPassword")} autoComplete="new-password" />
                  </div>
                )}
              </div>
            </div>
            <div className="flex justify-end">
              <button
                type="submit"
                className="mt-4 w-auto flex items-center py-2 px-4 transition rounded-md shadow-xs text-sm font-medium text-white bg-blue-600 dark:bg-blue-600 hover:bg-blue-700 dark:hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-blue-500 cursor-pointer"
              >
                <UserIcon className="w-4 h-4 mr-1" />
                {t("account.save")}
              </button>
            </div>
          </form>
        </form.AppForm>
      </div>
    </Section>
  );
}

export default AccountSettings;
